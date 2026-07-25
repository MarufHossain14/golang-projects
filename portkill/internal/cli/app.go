// Package cli contains the command-line interface shared by every operating system.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/MarufHossain14/golang-projects/portkill/internal/process"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

type processFinder interface {
	FindByPort(port int) (process.Info, error)
}

// Run handles command-line arguments and returns a process exit code.
//
// Keeping this work outside main makes the CLI easy to test without starting
// another operating-system process.
func Run(args []string, stdout, stderr io.Writer, version string, finder processFinder) int {
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

	// Process discovery is implemented in the next milestone. Keeping this
	// message here lets us verify that valid options reached the right path.
	if options.List {
		fmt.Fprintln(stderr, "portkill: listing ports is not available yet")
		return exitFailure
	}

	info, err := finder.FindByPort(options.Port)
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
	fmt.Fprintln(stdout, "\nProcess discovery only; nothing was terminated.")

	return exitSuccess
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
