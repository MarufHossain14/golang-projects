package main

import "testing"

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
