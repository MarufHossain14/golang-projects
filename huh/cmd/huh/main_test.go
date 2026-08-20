package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarufHossain14/golang-projects/huh/internal/audio"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(context.Background(), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "instant replay") {
		t.Fatalf("help output = %q, want product description", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(context.Background(), []string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if got, want := stdout.String(), "huh "+version+"\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(context.Background(), []string{"wat"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "wat"`) {
		t.Fatalf("stderr = %q, want unknown-command error", stderr.String())
	}
}

func TestCaptureRequiresExplicitOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(context.Background(), []string{"capture"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "--output is required") {
		t.Fatalf("stderr = %q, want explicit-output error", stderr.String())
	}
}

func TestExportWAVRefusesToOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.wav")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := exportWAV(path, audio.SpeechPCM(), []byte{1, 2})

	if err == nil {
		t.Fatal("exportWAV() error = nil, want existing-file error")
	}
	contents, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if got, want := string(contents), "keep me"; got != want {
		t.Fatalf("existing contents = %q, want %q", got, want)
	}
}

func TestExportWAVUsesPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.wav")

	if err := exportWAV(path, audio.SpeechPCM(), []byte{1, 2}); err != nil {
		t.Fatalf("exportWAV() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("permissions = %o, want %o", got, want)
	}
}
