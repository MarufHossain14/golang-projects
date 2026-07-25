// Package cli contains portkill's command-line behavior.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/MarufHossain14/golang-projects/portkill/internal/process"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

type processManager interface {
	FindByPort(port int) (process.Info, error)
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
		fmt.Fprintln(stderr, "portkill: listing ports is not available yet")
		return exitFailure
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

	fmt.Fprintf(stdout, "Found process using port %d\n\n", info.Port)
	fmt.Fprintf(stdout, "Process: %s\n", info.Name)
	fmt.Fprintf(stdout, "PID: %d\n", info.PID)
	if info.Command != "" {
		fmt.Fprintf(stdout, "Command: %s\n", info.Command)
	}

	if options.DryRun {
		fmt.Fprintln(stdout, "\nDry run: this process would be terminated.")
		return exitSuccess
	}

	if !options.Force {
		confirmed, err := askForConfirmation(stdin, stdout)
		if err != nil {
			fmt.Fprintf(stderr, "portkill: confirmation: %v\n", err)
			return exitFailure
		}
		if !confirmed {
			fmt.Fprintln(stdout, "Cancelled. Process was not terminated.")
			return exitSuccess
		}
	}

	if err := manager.Terminate(info.PID); err != nil {
		fmt.Fprintf(stderr, "portkill: terminate process: %v\n", err)
		return exitFailure
	}

	fmt.Fprintln(stdout, "Successfully terminated process.")

	return exitSuccess
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
