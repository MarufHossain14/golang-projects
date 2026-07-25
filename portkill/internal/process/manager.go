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
	"syscall"
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

// Manager finds and terminates Linux processes.
type Manager struct {
	run      func(name string, args ...string) ([]byte, error)
	readFile func(name string) ([]byte, error)
	signal   func(pid int, signal os.Signal) error
}

// NewManager creates a Manager that uses the real operating system.
func NewManager() *Manager {
	return &Manager{
		run: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
		readFile: os.ReadFile,
		signal: func(pid int, signal os.Signal) error {
			target, err := os.FindProcess(pid)
			if err != nil {
				return err
			}
			return target.Signal(signal)
		},
	}
}

// FindByPort finds the process listening on a TCP port.
func (m *Manager) FindByPort(port int) (Info, error) {
	output, err := m.run(
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

	commandLine, err := m.readFile(fmt.Sprintf("/proc/%d/cmdline", info.PID))
	if err == nil {
		info.Command = parseCommandLine(commandLine)
	}

	return info, nil
}

// Terminate asks a process to exit cleanly using the Linux SIGTERM signal.
func (m *Manager) Terminate(pid int) error {
	if pid < 1 {
		return fmt.Errorf("PID must be greater than zero")
	}

	if err := m.signal(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal PID %d: %w", pid, err)
	}

	return nil
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
