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

func TestParseLsofList(t *testing.T) {
	output := []byte(
		"p12345\n" +
			"cnode\n" +
			"n*:3001\n" +
			"n127.0.0.1:3000\n" +
			"n[::]:3000\n" +
			"p9123\n" +
			"cpostgres\n" +
			"n127.0.0.1:5432\n",
	)

	got, err := parseLsofList(output)
	if err != nil {
		t.Fatalf("parseLsofList() returned an unexpected error: %v", err)
	}

	want := []Info{
		{Port: 3000, PID: 12345, Name: "node"},
		{Port: 3001, PID: 12345, Name: "node"},
		{Port: 5432, PID: 9123, Name: "postgres"},
	}
	if len(got) != len(want) {
		t.Fatalf("parseLsofList() returned %d processes, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("process %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestPortFromAddress(t *testing.T) {
	tests := []struct {
		address string
		want    int
	}{
		{address: "*:3000", want: 3000},
		{address: "127.0.0.1:8080", want: 8080},
		{address: "[::1]:5432", want: 5432},
	}

	for _, test := range tests {
		got, err := portFromAddress(test.address)
		if err != nil {
			t.Fatalf("portFromAddress(%q) returned an unexpected error: %v", test.address, err)
		}
		if got != test.want {
			t.Fatalf("portFromAddress(%q) = %d, want %d", test.address, got, test.want)
		}
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

func TestManagerList(t *testing.T) {
	manager := &Manager{
		run: func(string, ...string) ([]byte, error) {
			return []byte("p12345\ncnode\nn*:3000\n"), nil
		},
		readFile: func(string) ([]byte, error) {
			return []byte("pnpm\x00dev\x00"), nil
		},
	}

	got, err := manager.List()
	if err != nil {
		t.Fatalf("List() returned an unexpected error: %v", err)
	}

	want := []Info{{
		Port:    3000,
		PID:     12345,
		Name:    "node",
		Command: "pnpm dev",
	}}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("List() = %+v, want %+v", got, want)
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
