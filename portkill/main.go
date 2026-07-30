package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Release builds replace "dev" with a real version number.
var version = "dev"

// options keeps all values provided by the user in one place.
type options struct {
	port   int
	force  bool
	dryRun bool
	list   bool
	json   bool
}

func main() {
	os.Exit(run(os.Args[1:]))
}

// run handles the command and returns the exit code used by the terminal.
func run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}
	if len(args) == 1 {
		switch args[0] {
		case "help", "-h", "--help":
			printHelp()
			return 0
		case "version", "-v", "--version":
			fmt.Printf("portkill %s\n", version)
			return 0
		}
	}

	opts, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 2
	}

	if opts.list {
		processes, err := listProcesses()
		if err != nil {
			fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
			return 1
		}
		if opts.json {
			if err := json.NewEncoder(os.Stdout).Encode(processes); err != nil {
				fmt.Fprintf(os.Stderr, "portkill: failed to write JSON output: %v\n", err)
				return 1
			}
		} else {
			printTable(processes)
		}
		return 0
	}

	process, err := findProcess(opts.port)
	if errors.Is(err, errNotFound) {
		fmt.Fprintf(os.Stderr, "portkill: no process is listening on port %d\n", opts.port)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 1
	}

	if opts.json {
		if err := json.NewEncoder(os.Stdout).Encode(process); err != nil {
			fmt.Fprintf(os.Stderr, "portkill: failed to write JSON output: %v\n", err)
			return 1
		}
	} else {
		fmt.Printf("✓ Found process using port %d\n\n", process.Port)
		fmt.Printf("Process: %s\nPID: %d\nCommand: %s\n", process.Name, process.PID, process.Command)
	}

	status := os.Stdout
	if opts.json {
		// Keep stdout as valid JSON so another program can read it.
		status = os.Stderr
	}
	if opts.dryRun {
		fmt.Fprintln(status, "Dry run: this process would be terminated.")
		return 0
	}
	if !opts.force && !confirm(status) {
		fmt.Fprintln(status, "Cancelled. Process was not terminated.")
		return 0
	}
	if err := terminateProcess(process.PID); err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 1
	}
	fmt.Fprintln(status, "✓ Successfully terminated process.")
	return 0
}

// parseOptions converts terminal arguments into validated options.
func parseOptions(args []string) (options, error) {
	var opts options

	for _, arg := range args {
		switch arg {
		case "-f", "--force":
			opts.force = true
		case "-d", "--dry-run":
			opts.dryRun = true
		case "list", "-l", "--list":
			opts.list = true
		case "-j", "--json":
			opts.json = true
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("unknown option %q", arg)
			}
			if opts.port != 0 {
				return opts, fmt.Errorf("only one port can be provided")
			}
			port, err := strconv.Atoi(arg)
			if err != nil || port < 1 || port > 65535 {
				return opts, fmt.Errorf("port must be a number between 1 and 65535")
			}
			opts.port = port
		}
	}

	if opts.list && opts.port != 0 {
		return opts, fmt.Errorf("--list cannot be used with a port")
	}
	if opts.list && (opts.force || opts.dryRun) {
		return opts, fmt.Errorf("--force and --dry-run require a port")
	}
	if !opts.list && opts.port == 0 {
		return opts, fmt.Errorf("provide a port or use --list")
	}
	if opts.force && opts.dryRun {
		return opts, fmt.Errorf("--force and --dry-run cannot be used together")
	}
	return opts, nil
}

// confirm uses yes as the default when the user only presses Enter.
func confirm(output *os.File) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(output, "\nKill this process? (Y/n) ")
		answer, err := reader.ReadString('\n')
		if err != nil && answer == "" {
			return false
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Fprintln(output, "Please answer y or n.")
		}
	}
}

// printTable lines up the columns without needing an external package.
func printTable(processes []Process) {
	if len(processes) == 0 {
		fmt.Println("No listening TCP ports found.")
		return
	}
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "PORT\tPID\tPROCESS\tCOMMAND")
	for _, process := range processes {
		fmt.Fprintf(writer, "%d\t%d\t%s\t%s\n", process.Port, process.PID, process.Name, process.Command)
	}
	writer.Flush()
}

func printHelp() {
	fmt.Println(`portkill finds and terminates the process using a TCP port.

Usage:
  portkill <port> [options]
  portkill --list [--json]
  portkill list [--json]

Options:
  -f, --force     Skip confirmation
  -d, --dry-run   Show the process without terminating it
  -l, --list      List listening TCP ports
  -j, --json      Print JSON output
  -h, --help      Show this help
  -v, --version   Show the version

Examples:
  portkill 3000
  portkill 3000 --dry-run
  portkill 3000 --force
  portkill --list
  portkill list --json`)
}
