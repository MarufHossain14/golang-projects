package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--help"}, &stdout, &stderr, "dev")

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

	exitCode := Run([]string{"version"}, &stdout, &stderr, "1.2.3")

	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}
	if got, want := stdout.String(), "portkill 1.2.3\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRunUnknownArgument(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"3000"}, &stdout, &stderr, "dev")

	if exitCode != exitUsage {
		t.Fatalf("expected exit code %d, got %d", exitUsage, exitCode)
	}
	if !strings.Contains(stderr.String(), "not available yet") {
		t.Fatalf("expected a useful error, got %q", stderr.String())
	}
}
