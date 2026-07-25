package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MarufHossain14/golang-projects/portkill/internal/process"
)

type stubManager struct {
	info          process.Info
	findErr       error
	terminateErr  error
	terminatedPID *int
}

func (s stubManager) FindByPort(int) (process.Info, error) {
	return s.info, s.findErr
}

func (s stubManager) Terminate(pid int) error {
	if s.terminatedPID != nil {
		*s.terminatedPID = pid
	}
	return s.terminateErr
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr, "dev", stubManager{})

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no error output, got %q", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"version"}, strings.NewReader(""), &stdout, &stderr, "1.2.3", stubManager{})

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if got, want := stdout.String(), "portkill 1.2.3\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRunDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	manager := stubManager{
		info: process.Info{
			Port:    3000,
			PID:     12345,
			Name:    "node",
			Command: "pnpm dev",
		},
	}
	exitCode := Run(
		[]string{"3000", "--dry-run"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if !strings.Contains(stdout.String(), "Process: node") {
		t.Fatalf("expected process details, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "PID: 12345") {
		t.Fatalf("expected PID, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Dry run") {
		t.Fatalf("expected dry-run message, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no error output, got %q", stderr.String())
	}
}

func TestRunForceTerminatesProcess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var terminatedPID int

	manager := stubManager{
		info: process.Info{
			Port: 3000,
			PID:  12345,
			Name: "node",
		},
		terminatedPID: &terminatedPID,
	}
	exitCode := Run(
		[]string{"3000", "--force"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if terminatedPID != 12345 {
		t.Fatalf("terminated PID %d, want 12345", terminatedPID)
	}
	if !strings.Contains(stdout.String(), "Successfully terminated") {
		t.Fatalf("expected success output, got %q", stdout.String())
	}
}

func TestRunCanCancelTermination(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var terminatedPID int

	manager := stubManager{
		info: process.Info{
			Port: 3000,
			PID:  12345,
			Name: "node",
		},
		terminatedPID: &terminatedPID,
	}
	exitCode := Run(
		[]string{"3000"},
		strings.NewReader("n\n"),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if terminatedPID != 0 {
		t.Fatalf("expected no terminated process, got PID %d", terminatedPID)
	}
	if !strings.Contains(stdout.String(), "Cancelled") {
		t.Fatalf("expected cancellation output, got %q", stdout.String())
	}
}

func TestAskForConfirmation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{name: "enter uses yes default", input: "\n", wantOK: true},
		{name: "yes", input: "yes\n", wantOK: true},
		{name: "no", input: "n\n", wantOK: false},
		{name: "asks again", input: "maybe\ny\n", wantOK: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer

			got, err := askForConfirmation(strings.NewReader(test.input), &output)
			if err != nil {
				t.Fatalf("askForConfirmation() returned an unexpected error: %v", err)
			}
			if got != test.wantOK {
				t.Fatalf("askForConfirmation() = %t, want %t", got, test.wantOK)
			}
		})
	}
}

func TestRunProcessNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	manager := stubManager{findErr: process.ErrNotFound}
	exitCode := Run(
		[]string{"3000"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitFailure {
		t.Fatalf("expected exit code %d, got %d", exitFailure, exitCode)
	}
	if !strings.Contains(stderr.String(), "no process is listening on port 3000") {
		t.Fatalf("expected a useful error, got %q", stderr.String())
	}
}

func TestRunFinderError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	manager := stubManager{findErr: errors.New("lsof failed")}
	exitCode := Run(
		[]string{"3000"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitFailure {
		t.Fatalf("expected exit code %d, got %d", exitFailure, exitCode)
	}
	if !strings.Contains(stderr.String(), "find process: lsof failed") {
		t.Fatalf("expected a useful error, got %q", stderr.String())
	}
}

func TestRunTerminateError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	manager := stubManager{
		info: process.Info{
			Port: 3000,
			PID:  12345,
			Name: "node",
		},
		terminateErr: errors.New("permission denied"),
	}
	exitCode := Run(
		[]string{"3000", "--force"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"dev",
		manager,
	)

	if exitCode != exitFailure {
		t.Fatalf("expected exit code %d, got %d", exitFailure, exitCode)
	}
	if !strings.Contains(stderr.String(), "permission denied") {
		t.Fatalf("expected termination error, got %q", stderr.String())
	}
}
