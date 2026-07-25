package process

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestParseLsofOutput(t *testing.T) {
	output := []byte("p12345\ncnode\n")

	got, err := parseLsofOutput(output, 3000)
	if err != nil {
		t.Fatalf("parseLsofOutput() returned an unexpected error: %v", err)
	}

	want := Info{Port: 3000, PID: 12345, Name: "node"}
	if got != want {
		t.Fatalf("parseLsofOutput() = %+v, want %+v", got, want)
	}
}

func TestParseLsofOutputWithoutProcess(t *testing.T) {
	_, err := parseLsofOutput(nil, 3000)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestParseCommandLine(t *testing.T) {
	got := parseCommandLine([]byte("node\x00server.js\x00--port\x003000\x00"))
	want := "node server.js --port 3000"

	if got != want {
		t.Fatalf("parseCommandLine() = %q, want %q", got, want)
	}
}

func TestManagerFindByPort(t *testing.T) {
	manager := &Manager{
		run: func(string, ...string) ([]byte, error) {
			return []byte("p12345\ncnode\n"), nil
		},
		readFile: func(string) ([]byte, error) {
			return []byte("pnpm\x00dev\x00"), nil
		},
	}

	got, err := manager.FindByPort(3000)
	if err != nil {
		t.Fatalf("FindByPort() returned an unexpected error: %v", err)
	}

	want := Info{
		Port:    3000,
		PID:     12345,
		Name:    "node",
		Command: "pnpm dev",
	}
	if got != want {
		t.Fatalf("FindByPort() = %+v, want %+v", got, want)
	}
}

func TestManagerTerminate(t *testing.T) {
	var gotPID int
	var gotSignal os.Signal
	manager := &Manager{
		signal: func(pid int, signal os.Signal) error {
			gotPID = pid
			gotSignal = signal
			return nil
		},
	}

	if err := manager.Terminate(12345); err != nil {
		t.Fatalf("Terminate() returned an unexpected error: %v", err)
	}
	if gotPID != 12345 {
		t.Fatalf("Terminate() signalled PID %d, want 12345", gotPID)
	}
	if gotSignal != syscall.SIGTERM {
		t.Fatalf("Terminate() used signal %v, want %v", gotSignal, syscall.SIGTERM)
	}
}

func TestManagerTerminateRejectsInvalidPID(t *testing.T) {
	manager := &Manager{}

	if err := manager.Terminate(0); err == nil {
		t.Fatal("Terminate() returned no error for PID 0")
	}
}
