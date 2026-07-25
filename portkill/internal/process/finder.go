// Package process finds information about Linux processes using network ports.
package process

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ErrNotFound means that no process is listening on the requested port.
var ErrNotFound = errors.New("no listening process found")

// Info contains the details portkill needs about a process.
type Info struct {
	Port    int
	PID     int
	Name    string
	Command string
}

// Finder uses Linux tools and the /proc filesystem to find processes.
type Finder struct {
	run      func(name string, args ...string) ([]byte, error)
	readFile func(name string) ([]byte, error)
}

// NewFinder creates a Finder that uses the real operating system.
func NewFinder() *Finder {
	return &Finder{
		run: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
		readFile: os.ReadFile,
	}
}

// FindByPort finds the process listening on a TCP port.
func (f *Finder) FindByPort(port int) (Info, error) {
	output, err := f.run(
		"lsof",
		"-nP",
		"-a",
		fmt.Sprintf("-iTCP:%d", port),
		"-sTCP:LISTEN",
		"-Fpc",
	)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Info{}, fmt.Errorf("lsof is required but was not found")
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
			return Info{}, fmt.Errorf("%w on port %d", ErrNotFound, port)
		}

		return Info{}, fmt.Errorf("run lsof: %w", err)
	}

	info, err := parseLsofOutput(output, port)
	if err != nil {
		return Info{}, err
	}

	commandLine, err := f.readFile(fmt.Sprintf("/proc/%d/cmdline", info.PID))
	if err == nil {
		info.Command = parseCommandLine(commandLine)
	}

	return info, nil
}

func parseLsofOutput(output []byte, port int) (Info, error) {
	info := Info{Port: port}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(line[1:])
			if err != nil {
				return Info{}, fmt.Errorf("read PID from lsof: %w", err)
			}
			info.PID = pid
		case 'c':
			info.Name = line[1:]
		}

		if info.PID != 0 && info.Name != "" {
			return info, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return Info{}, fmt.Errorf("read lsof output: %w", err)
	}

	return Info{}, fmt.Errorf("%w on port %d", ErrNotFound, port)
}

func parseCommandLine(commandLine []byte) string {
	parts := bytes.Split(commandLine, []byte{0})
	arguments := make([]string, 0, len(parts))

	for _, part := range parts {
		if argument := strings.TrimSpace(string(part)); argument != "" {
			arguments = append(arguments, argument)
		}
	}

	return strings.Join(arguments, " ")
}
