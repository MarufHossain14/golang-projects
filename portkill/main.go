package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type options struct {
	port  int
	force bool
}

func main() {
	os.Exit(run(os.Args[1:]))
}

// run handles the command and returns the exit code used by the terminal.
func run(args []string) int {
	if len(args) == 0 {
		processes, err := listProcesses()
		if err != nil {
			fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
			return 1
		}
		input := bufio.NewReader(os.Stdin)
		process, ok := chooseProcess(processes, input, os.Stdout)
		if !ok {
			return 0
		}
		details := processInfo(process.PID, process.Port)
		if details.Name == "" {
			details.Name = process.Name
		}
		return stopProcess(details, false, input)
	}
	if len(args) == 1 && args[0] == "list" {
		return showProcesses()
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printHelp()
		return 0
	}

	opts, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 2
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

	return stopProcess(process, opts.force, os.Stdin)
}

func stopProcess(process Process, force bool, input io.Reader) int {
	fmt.Printf("\nPort %d is used by %s (PID %d)\n", process.Port, process.Name, process.PID)
	if process.Command != "" {
		fmt.Printf("Command: %s\n", process.Command)
	}

	if !force && !confirm(input, os.Stdout) {
		fmt.Println("Cancelled.")
		return 0
	}
	if err := verifyProcess(process.Port, process.PID); err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 1
	}
	if err := terminateProcess(process.PID); err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 1
	}
	fmt.Printf("Sent SIGTERM to PID %d.\n", process.PID)
	return 0
}

func chooseProcess(processes []Process, input io.Reader, output io.Writer) (Process, bool) {
	if len(processes) == 0 {
		fmt.Fprintln(output, "No killable listening TCP ports found.")
		return Process{}, false
	}

	fmt.Fprintf(output, "%-4s %-7s %-8s %s\n", "#", "PORT", "PID", "PROCESS")
	for i, process := range processes {
		fmt.Fprintf(output, "%-4d %-7d %-8d %s\n", i+1, process.Port, process.PID, process.Name)
	}

	reader := bufferedReader(input)
	for {
		fmt.Fprintf(output, "\nSelect a process (1-%d, Enter to cancel): ", len(processes))
		answer, err := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer == "" {
			fmt.Fprintln(output, "Cancelled.")
			return Process{}, false
		}
		selection, parseErr := strconv.Atoi(answer)
		if parseErr == nil && selection >= 1 && selection <= len(processes) {
			return processes[selection-1], true
		}
		if err != nil {
			fmt.Fprintln(output, "Cancelled.")
			return Process{}, false
		}
		fmt.Fprintln(output, "Please enter a number from the list.")
	}
}

func showProcesses() int {
	processes, err := listProcesses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "portkill: %v\n", err)
		return 1
	}
	if len(processes) == 0 {
		fmt.Println("No killable listening TCP ports found.")
		return 0
	}

	fmt.Printf("%-7s %-8s %s\n", "PORT", "PID", "PROCESS")
	for _, process := range processes {
		fmt.Printf("%-7d %-8d %s\n", process.Port, process.PID, process.Name)
	}
	return 0
}

func parseOptions(args []string) (options, error) {
	var opts options

	for _, arg := range args {
		switch arg {
		case "-f", "--force":
			opts.force = true
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

	if opts.port == 0 {
		return opts, fmt.Errorf("provide a port")
	}
	return opts, nil
}

// confirm requires an explicit yes so an accidental Enter cannot kill a process.
func confirm(input io.Reader, output io.Writer) bool {
	reader := bufferedReader(input)
	for {
		fmt.Fprint(output, "\nKill this process? (y/N) ")
		answer, err := reader.ReadString('\n')
		if err != nil && answer == "" {
			return false
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "y", "yes":
			return true
		case "", "n", "no":
			return false
		default:
			fmt.Fprintln(output, "Please answer y or n.")
		}
	}
}

// bufferedReader preserves any input already buffered by a previous prompt.
func bufferedReader(input io.Reader) *bufio.Reader {
	if reader, ok := input.(*bufio.Reader); ok {
		return reader
	}
	return bufio.NewReader(input)
}

func printHelp() {
	fmt.Println(`portkill finds and terminates the process using a TCP port.

Usage:
  portkill
  portkill list
  portkill <port> [--force]

Options:
  -f, --force  Skip confirmation
  -h, --help   Show this help

Examples:
  portkill              Select a process from a list
  portkill list         Only show the list
  portkill 3000
  portkill 3000 --force`)
}
