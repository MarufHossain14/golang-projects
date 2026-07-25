package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MarufHossain14/golang-projects/portkill/internal/process"
)

type stubFinder struct {
	info process.Info
	err  error
}

func (s stubFinder) FindByPort(int) (process.Info, error) {
	return s.info, s.err
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--help"}, &stdout, &stderr, "dev", stubFinder{})

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

	exitCode := Run([]string{"version"}, &stdout, &stderr, "1.2.3", stubFinder{})

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if got, want := stdout.String(), "portkill 1.2.3\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRunFindsProcess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	finder := stubFinder{
		info: process.Info{
			Port:    3000,
			PID:     12345,
			Name:    "node",
			Command: "pnpm dev",
		},
	}
	exitCode := Run([]string{"3000"}, &stdout, &stderr, "dev", finder)

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if !strings.Contains(stdout.String(), "Process: node") {
		t.Fatalf("expected process details, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "PID: 12345") {
		t.Fatalf("expected PID, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no error output, got %q", stderr.String())
	}
}

func TestRunProcessNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	finder := stubFinder{err: process.ErrNotFound}
	exitCode := Run([]string{"3000"}, &stdout, &stderr, "dev", finder)

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

	finder := stubFinder{err: errors.New("lsof failed")}
	exitCode := Run([]string{"3000"}, &stdout, &stderr, "dev", finder)

	if exitCode != exitFailure {
		t.Fatalf("expected exit code %d, got %d", exitFailure, exitCode)
	}
	if !strings.Contains(stderr.String(), "find process: lsof failed") {
		t.Fatalf("expected a useful error, got %q", stderr.String())
	}
}
