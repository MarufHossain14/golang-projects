package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// Options contains the user's requested portkill operation.
type Options struct {
	Port   int
	Force  bool
	DryRun bool
	List   bool
	JSON   bool
}

// ParseOptions turns command-line arguments into validated options.
func ParseOptions(args []string) (Options, error) {
	var options Options
	var portArgument string

	for _, argument := range args {
		switch argument {
		case "--force":
			options.Force = true
		case "--dry-run":
			options.DryRun = true
		case "--list":
			options.List = true
		case "--json":
			options.JSON = true
		default:
			if strings.HasPrefix(argument, "-") {
				return Options{}, fmt.Errorf("unknown option %q", argument)
			}
			if portArgument != "" {
				return Options{}, fmt.Errorf("expected one port, received %q and %q", portArgument, argument)
			}
			portArgument = argument
		}
	}

	if options.List {
		if portArgument != "" {
			return Options{}, fmt.Errorf("--list cannot be used with a port")
		}
		if options.Force || options.DryRun {
			return Options{}, fmt.Errorf("--force and --dry-run require a port")
		}
		return options, nil
	}

	if portArgument == "" {
		return Options{}, fmt.Errorf("provide a port or use --list")
	}
	if options.Force && options.DryRun {
		return Options{}, fmt.Errorf("--force and --dry-run cannot be used together")
	}

	port, err := strconv.Atoi(portArgument)
	if err != nil {
		return Options{}, fmt.Errorf("port must be a number, got %q", portArgument)
	}
	if port < 1 || port > 65535 {
		return Options{}, fmt.Errorf("port must be between 1 and 65535")
	}

	options.Port = port
	return options, nil
}
