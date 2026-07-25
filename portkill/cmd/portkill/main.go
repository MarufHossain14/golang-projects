// Command portkill finds and terminates the process listening on a network port.
package main

import (
	"os"

	"github.com/MarufHossain14/golang-projects/portkill/internal/cli"
)

// version is replaced with a release version when the binary is built.
var version = "dev"

func main() {
	exitCode := cli.Run(os.Args[1:], os.Stdout, os.Stderr, version)
	os.Exit(exitCode)
}
