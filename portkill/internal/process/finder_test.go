package process

import (
	"errors"
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

func TestFinderFindByPort(t *testing.T) {
	finder := &Finder{
		run: func(string, ...string) ([]byte, error) {
			return []byte("p12345\ncnode\n"), nil
		},
		readFile: func(string) ([]byte, error) {
			return []byte("pnpm\x00dev\x00"), nil
		},
	}

	got, err := finder.FindByPort(3000)
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
