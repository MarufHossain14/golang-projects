package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestParseOptions(t *testing.T) {
	options, err := parseOptions([]string{"3000", "--force"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if options.port != 3000 || !options.force {
		t.Fatalf("unexpected options: %+v", options)
	}

	if _, err := parseOptions([]string{"70000"}); err == nil {
		t.Fatal("expected an invalid port error")
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

func TestChooseProcess(t *testing.T) {
	processes := []Process{
		{Port: 3000, PID: 8014, Name: "python3"},
		{Port: 8080, PID: 9000, Name: "server"},
	}
	var output bytes.Buffer

	process, ok := chooseProcess(processes, strings.NewReader("wrong\n2\n"), &output)
	if !ok {
		t.Fatal("expected a process to be selected")
	}
	if process.Port != 8080 || process.PID != 9000 {
		t.Fatalf("unexpected process: %+v", process)
	}
	if !strings.Contains(output.String(), "Please enter a number from the list.") {
		t.Fatalf("expected invalid selection message, got %q", output.String())
	}
}

func TestChooseThenConfirmUsesSameInput(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("1\ny\n"))
	var output bytes.Buffer
	processes := []Process{{Port: 3000, PID: 8014, Name: "python3"}}

	if _, ok := chooseProcess(processes, input, &output); !ok {
		t.Fatal("expected a process to be selected")
	}
	if !confirm(input, &output) {
		t.Fatal("expected confirmation after selection")
	}
}

func TestParseSSLine(t *testing.T) {
	process, err := parseSSLine(
		`LISTEN 0 511 *:3000 *:* users:(("node",pid=12345,fd=20))`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if process.Port != 3000 || process.PID != 12345 || process.Name != "node" {
		t.Fatalf("unexpected process: %+v", process)
	}
}

func TestParseSSLineWithoutPID(t *testing.T) {
	_, err := parseSSLine("LISTEN 0 511 *:3000 *:*")
	if err == nil {
		t.Fatal("expected missing PID to fail")
	}
}

func TestParseSSProcesses(t *testing.T) {
	output := []byte(
		"LISTEN 0 5 0.0.0.0:3000 0.0.0.0:* users:((\"python3\",pid=8014,fd=3))\n" +
			"LISTEN 0 5 [::]:3000 [::]:* users:((\"python3\",pid=8014,fd=4))\n" +
			"LISTEN 0 1000 10.255.255.254:53 0.0.0.0:*\n",
	)

	processes := parseSSProcesses(output)
	if len(processes) != 1 {
		t.Fatalf("got %d processes, want 1", len(processes))
	}
	if processes[0].Port != 3000 || processes[0].PID != 8014 || processes[0].Name != "python3" {
		t.Fatalf("unexpected process: %+v", processes[0])
	}
}
