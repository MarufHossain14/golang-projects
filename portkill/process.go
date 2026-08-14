package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

var errNotFound = errors.New("no listening process found")

// Process contains the information shown to the user.
type Process struct {
	Port    int
	PID     int
	Name    string
	Command string
}

// findProcess asks ss for the PID listening on one TCP port.
func findProcess(port int) (Process, error) {
	output, err := exec.Command(
		"ss", "-H", "-ltnp",
		fmt.Sprintf("sport = :%d", port),
	).Output()
	if err != nil {
		return Process{}, handleSSError(err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return Process{}, errNotFound
	}
	found, err := parseSSLine(lines[0])
	if err != nil {
		return Process{}, err
	}

	info := processInfo(found.PID, found.Port)
	if info.Name == "" {
		info.Name = found.Name
	}
	if info.Name == "" {
		info.Name = "unknown process"
	}
	return info, nil
}

// listProcesses returns listeners that expose a PID and can therefore be killed.
func listProcesses() ([]Process, error) {
	output, err := exec.Command("ss", "-H", "-ltnp").Output()
	if err != nil {
		return nil, handleSSError(err)
	}
	return parseSSProcesses(output), nil
}

func parseSSProcesses(output []byte) []Process {
	processes := make([]Process, 0)
	seen := make(map[[2]int]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		process, err := parseSSLine(line)
		if err != nil {
			continue
		}
		key := [2]int{process.Port, process.PID}
		if seen[key] {
			continue
		}
		seen[key] = true
		if process.Name == "" {
			process.Name = "unknown process"
		}
		processes = append(processes, process)
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].Port < processes[j].Port
	})
	return processes
}

func parseSSLine(line string) (Process, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return Process{}, fmt.Errorf("could not read ss output")
	}

	port, err := portFromAddress(fields[3])
	if err != nil {
		return Process{}, err
	}

	pidStart := strings.Index(line, "pid=")
	if pidStart == -1 {
		return Process{}, fmt.Errorf("PID is not available")
	}
	pidText := line[pidStart+4:]
	pidEnd := strings.IndexFunc(pidText, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if pidEnd != -1 {
		pidText = pidText[:pidEnd]
	}
	pid, err := strconv.Atoi(pidText)
	if err != nil {
		return Process{}, fmt.Errorf("could not read PID from ss")
	}

	name := ""
	nameStart := strings.Index(line, `(("`)
	nameEnd := strings.Index(line, `",pid=`)
	if nameStart != -1 && nameEnd > nameStart {
		name = line[nameStart+3 : nameEnd]
	}

	return Process{Port: port, PID: pid, Name: name}, nil
}

func portFromAddress(address string) (int, error) {
	colon := strings.LastIndex(address, ":")
	if colon == -1 {
		return 0, fmt.Errorf("port not found")
	}
	port, err := strconv.Atoi(address[colon+1:])
	if err != nil {
		return 0, fmt.Errorf("invalid port in address %q", address)
	}
	return port, nil
}

// processInfo reads Linux's /proc folder for the name and full command.
func processInfo(pid, port int) Process {
	name, _ := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	return Process{
		Port:    port,
		PID:     pid,
		Name:    strings.TrimSpace(string(name)),
		Command: readCommand(pid),
	}
}

func readCommand(pid int) string {
	command, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	// Linux separates command arguments with null bytes instead of spaces.
	return strings.TrimSpace(string(bytes.ReplaceAll(command, []byte{0}, []byte(" "))))
}

// verifyProcess makes sure the PID found before confirmation still owns the port.
// This avoids signalling a different process if the original listener exited.
func verifyProcess(port, pid int) error {
	current, err := findProcess(port)
	if errors.Is(err, errNotFound) {
		return fmt.Errorf("nothing is listening on port %d anymore", port)
	}
	if err != nil {
		return fmt.Errorf("could not verify port %d before terminating: %w", port, err)
	}
	if current.PID != pid {
		return fmt.Errorf("port %d now belongs to PID %d; refusing to terminate it", port, current.PID)
	}
	return nil
}

// terminateProcess sends SIGTERM so the process can shut down cleanly.
func terminateProcess(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return fmt.Errorf("PID %d exited before it could be terminated", pid)
		}
		if errors.Is(err, syscall.EPERM) {
			return fmt.Errorf("permission denied terminating PID %d (try running with sudo)", pid)
		}
		return fmt.Errorf("could not terminate PID %d: %w", pid, err)
	}
	return nil
}

func handleSSError(err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("ss is required; install it with 'sudo apt install iproute2'")
	}
	return fmt.Errorf("ss failed: %w", err)
}
