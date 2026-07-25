// Package cli contains portkill's command-line behavior.
package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/MarufHossain14/golang-projects/portkill/internal/process"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

type processManager interface {
	FindByPort(port int) (process.Info, error)
	List() ([]process.Info, error)
	Terminate(pid int) error
}

// Run handles command-line arguments and returns a process exit code.
//
// Keeping this work outside main makes the CLI easy to test without starting
// another operating-system process.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer, version string, manager processManager) int {
	if len(args) == 0 {
		printHelp(stdout)
		return exitSuccess
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return exitSuccess
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "portkill %s\n", version)
		return exitSuccess
	}

	options, err := ParseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "portkill: %v\n", err)
		fmt.Fprintln(stderr, "Run 'portkill --help' for usage.")
		return exitUsage
	}

	if options.List {
		return runList(stdout, stderr, manager, options.JSON)
	}

	info, err := manager.FindByPort(options.Port)
	if errors.Is(err, process.ErrNotFound) {
		fmt.Fprintf(stderr, "portkill: no process is listening on port %d\n", options.Port)
		return exitFailure
	}
	if err != nil {
		fmt.Fprintf(stderr, "portkill: find process: %v\n", err)
		return exitFailure
	}

	if err := writeProcess(stdout, info, options.JSON); err != nil {
		fmt.Fprintf(stderr, "portkill: write output: %v\n", err)
		return exitFailure
	}

	statusOutput := stdout
	if options.JSON {
		statusOutput = stderr
	}

	if options.DryRun {
		fmt.Fprintln(statusOutput, "\nDry run: this process would be terminated.")
		return exitSuccess
	}

	if !options.Force {
		confirmed, err := askForConfirmation(stdin, statusOutput)
		if err != nil {
			fmt.Fprintf(stderr, "portkill: confirmation: %v\n", err)
			return exitFailure
		}
		if !confirmed {
			fmt.Fprintln(statusOutput, "Cancelled. Process was not terminated.")
			return exitSuccess
		}
	}

	if err := manager.Terminate(info.PID); err != nil {
		fmt.Fprintf(stderr, "portkill: terminate process: %v\n", err)
		return exitFailure
	}

	fmt.Fprintln(statusOutput, "Successfully terminated process.")

	return exitSuccess
}

func runList(stdout, stderr io.Writer, manager processManager, jsonOutput bool) int {
	processes, err := manager.List()
	if err != nil {
		fmt.Fprintf(stderr, "portkill: list processes: %v\n", err)
		return exitFailure
	}

	if jsonOutput {
		if processes == nil {
			processes = []process.Info{}
		}
		if err := json.NewEncoder(stdout).Encode(processes); err != nil {
			fmt.Fprintf(stderr, "portkill: write JSON: %v\n", err)
			return exitFailure
		}
		return exitSuccess
	}

	if len(processes) == 0 {
		fmt.Fprintln(stdout, "No listening TCP ports found.")
		return exitSuccess
	}

	writer := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "PORT\tPID\tPROCESS\tCOMMAND")
	for _, info := range processes {
		name := info.Name
		if name == "" {
			name = "-"
		}
		command := info.Command
		if command == "" {
			command = "-"
		}
		fmt.Fprintf(
			writer,
			"%d\t%d\t%s\t%s\n",
			info.Port,
			info.PID,
			name,
			strings.ReplaceAll(command, "\t", " "),
		)
	}
	if err := writer.Flush(); err != nil {
		fmt.Fprintf(stderr, "portkill: write table: %v\n", err)
		return exitFailure
	}

	return exitSuccess
}

func writeProcess(output io.Writer, info process.Info, jsonOutput bool) error {
	if jsonOutput {
		return json.NewEncoder(output).Encode(info)
	}

	fmt.Fprintf(output, "Found process using port %d\n\n", info.Port)
	fmt.Fprintf(output, "Process: %s\n", info.Name)
	fmt.Fprintf(output, "PID: %d\n", info.PID)
	if info.Command != "" {
		fmt.Fprintf(output, "Command: %s\n", info.Command)
	}
	return nil
}

func askForConfirmation(input io.Reader, output io.Writer) (bool, error) {
	reader := bufio.NewReader(input)

	for {
		fmt.Fprint(output, "\nKill this process? (Y/n) ")

		answer, err := reader.ReadString('\n')
		if err != nil && !(errors.Is(err, io.EOF) && answer != "") {
			return false, err
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(output, "Please answer y or n.")
		}

		if errors.Is(err, io.EOF) {
			return false, io.EOF
		}
	}
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `portkill finds and terminates the process using a network port.

Usage:
  portkill <port>
  portkill --list
  portkill version

Options:
  -h, --help      Show this help message
  -v, --version   Show the installed version
      --force     Skip the confirmation prompt
      --dry-run   Show what would be killed
      --json      Print machine-readable JSON
`)
}
