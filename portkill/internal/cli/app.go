// Package cli contains the command-line interface shared by every operating system.
package cli

import (
	"fmt"
	"io"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

// Run handles command-line arguments and returns a process exit code.
//
// Keeping this work outside main makes the CLI easy to test without starting
// another operating-system process.
func Run(args []string, stdout, stderr io.Writer, version string) int {
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

	fmt.Fprintf(stderr, "portkill: process lookup for port %d is not available yet\n", options.Port)
	return exitFailure
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
