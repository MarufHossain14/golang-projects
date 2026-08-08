package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseOptions(t *testing.T) {
	options, err := parseOptions([]string{"3000", "-d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if options.port != 3000 || !options.dryRun {
		t.Fatalf("unexpected options: %+v", options)
	}

	if _, err := parseOptions([]string{"70000"}); err == nil {
		t.Fatal("expected an invalid port error")
	}
}

func TestParseOptionsListCommand(t *testing.T) {
	options, err := parseOptions([]string{"list", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !options.list || !options.json {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestParseOptionsRejectsUnsafeJSONPrompt(t *testing.T) {
	_, err := parseOptions([]string{"3000", "--json"})
	if err == nil {
		t.Fatal("expected --json without --force or --dry-run to fail")
	}
}

func TestConfirmRequiresExplicitYes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "yes", input: "yes\n", want: true},
		{name: "short yes", input: "Y\n", want: true},
		{name: "empty defaults to no", input: "\n", want: false},
		{name: "no", input: "n\n", want: false},
		{name: "end of input", input: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			if got := confirm(strings.NewReader(tt.input), &output); got != tt.want {
				t.Fatalf("confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmRetriesInvalidAnswer(t *testing.T) {
	var output bytes.Buffer
	got := confirm(strings.NewReader("maybe\ny\n"), &output)
	if !got {
		t.Fatal("expected confirmation after a valid second answer")
	}
	if !strings.Contains(output.String(), "Please answer y or n.") {
		t.Fatalf("expected retry message, got %q", output.String())
	}
}

func TestParseSSList(t *testing.T) {
	output := []byte(
		"LISTEN 0 511 *:3000 *:* users:((\"node\",pid=12345,fd=20))\n" +
			"LISTEN 0 244 127.0.0.1:5432 0.0.0.0:* users:((\"postgres\",pid=9123,fd=5))\n",
	)

	processes := parseSSList(output)
	if len(processes) != 2 {
		t.Fatalf("got %d processes, want 2", len(processes))
	}
	if processes[0].Port != 3000 || processes[0].PID != 12345 {
		t.Fatalf("unexpected first process: %+v", processes[0])
	}
	if processes[1].Port != 5432 || processes[1].Name != "postgres" {
		t.Fatalf("unexpected second process: %+v", processes[1])
	}
}

func TestParseSSListDeduplicatesIPv4AndIPv6(t *testing.T) {
	output := []byte(
		"LISTEN 0 511 0.0.0.0:3000 0.0.0.0:* users:((\"node\",pid=12345,fd=20))\n" +
			"LISTEN 0 511 [::]:3000 [::]:* users:((\"node\",pid=12345,fd=21))\n",
	)

	processes := parseSSList(output)
	if len(processes) != 1 {
		t.Fatalf("got %d processes, want 1", len(processes))
	}
}

func TestParseSSListSortsMatchingPortsByPID(t *testing.T) {
	output := []byte(
		"LISTEN 0 511 *:3000 *:* users:((\"worker-b\",pid=200,fd=20))\n" +
			"LISTEN 0 511 *:3000 *:* users:((\"worker-a\",pid=100,fd=21))\n",
	)

	processes := parseSSList(output)
	if len(processes) != 2 {
		t.Fatalf("got %d processes, want 2", len(processes))
	}
	if processes[0].PID != 100 || processes[1].PID != 200 {
		t.Fatalf("processes are not sorted by PID: %+v", processes)
	}
}

func TestParseSSLineWithoutPID(t *testing.T) {
	_, err := parseSSLine("LISTEN 0 511 *:3000 *:*")
	if err == nil {
		t.Fatal("expected missing PID to fail")
	}
}
