package cli

import (
	"strings"
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Options
	}{
		{
			name: "port",
			args: []string{"3000"},
			want: Options{Port: 3000},
		},
		{
			name: "port with force",
			args: []string{"3000", "--force"},
			want: Options{Port: 3000, Force: true},
		},
		{
			name: "flags before port",
			args: []string{"--json", "--dry-run", "8080"},
			want: Options{Port: 8080, DryRun: true, JSON: true},
		},
		{
			name: "list as JSON",
			args: []string{"--list", "--json"},
			want: Options{List: true, JSON: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseOptions(test.args)
			if err != nil {
				t.Fatalf("ParseOptions() returned an unexpected error: %v", err)
			}
			if got != test.want {
				t.Fatalf("ParseOptions() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseOptionsRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{name: "missing operation", args: nil, wantMessage: "provide a port"},
		{name: "non-numeric port", args: []string{"hello"}, wantMessage: "port must be a number"},
		{name: "zero port", args: []string{"0"}, wantMessage: "between 1 and 65535"},
		{name: "port too large", args: []string{"65536"}, wantMessage: "between 1 and 65535"},
		{name: "multiple ports", args: []string{"3000", "4000"}, wantMessage: "expected one port"},
		{name: "unknown option", args: []string{"3000", "--fast"}, wantMessage: "unknown option"},
		{name: "list with port", args: []string{"3000", "--list"}, wantMessage: "--list cannot be used with a port"},
		{name: "list with force", args: []string{"--list", "--force"}, wantMessage: "require a port"},
		{name: "conflicting modes", args: []string{"3000", "--force", "--dry-run"}, wantMessage: "cannot be used together"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseOptions(test.args)
			if err == nil {
				t.Fatal("ParseOptions() returned no error")
			}
			if !strings.Contains(err.Error(), test.wantMessage) {
				t.Fatalf("error %q does not contain %q", err, test.wantMessage)
			}
		})
	}
}
