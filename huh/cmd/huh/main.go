package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/MarufHossain14/golang-projects/huh/internal/audio"
	"github.com/MarufHossain14/golang-projects/huh/internal/buffer"
	"github.com/MarufHossain14/golang-projects/huh/internal/pipewire"
	"github.com/MarufHossain14/golang-projects/huh/internal/session"
)

const version = "0.1.0-dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintf(stdout, "huh %s\n", version)
		return 0
	case "doctor":
		return runDoctor(stdout)
	case "capture":
		return runCapture(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "huh: unknown command %q\n", args[0])
		fmt.Fprintln(stderr, "Run 'huh help' to see available commands.")
		return 2
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `HUH? — instant replay for the sentence you just missed

Usage:
  huh <command>

Commands:
  capture   Capture a rolling buffer and export it when stopped (development)
  doctor    Check whether planned audio and transcription tools are available
  version   Print the development version
  help      Show this help

The capture command is an explicit development tool. The normal listening mode
will keep audio memory-only and will not create recordings.`)
}

func runCapture(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("capture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	duration := flags.Duration("duration", 30*time.Second, "rolling buffer duration")
	output := flags.String("output", "", "destination WAV file (required)")
	target := flags.String("target", "", "PipeWire target node (default: automatic)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "huh capture: positional arguments are not supported")
		return 2
	}
	if *output == "" {
		fmt.Fprintln(stderr, "huh capture: --output is required because audio is never saved implicitly")
		return 2
	}
	if *duration < time.Second || *duration > time.Minute {
		fmt.Fprintln(stderr, "huh capture: --duration must be between 1s and 1m")
		return 2
	}

	format := audio.SpeechPCM()
	capacity, err := format.BytesFor(*duration)
	if err != nil {
		fmt.Fprintf(stderr, "huh capture: %v\n", err)
		return 1
	}
	ring, _ := buffer.New(capacity)
	recorder, _ := session.NewRecorder(ring)

	captureCtx, cancel := context.WithCancel(context.Background())
	source, err := pipewire.Open(captureCtx, *target)
	if err != nil {
		cancel()
		fmt.Fprintf(stderr, "huh capture: %v\n", err)
		return 1
	}

	done := make(chan error, 1)
	go func() {
		done <- recorder.Run(captureCtx, source)
	}()

	fmt.Fprintf(stdout, "Listening with a %s rolling buffer. Press Ctrl+C to export it.\n", duration.String())
	select {
	case err := <-done:
		cancel()
		if err != nil {
			fmt.Fprintf(stderr, "huh capture: %v\n", err)
			return 1
		}
		fmt.Fprintln(stderr, "huh capture: audio source stopped before export")
		return 1
	case <-ctx.Done():
		// Snapshot before cancellation because Recorder deliberately scrubs its
		// memory as part of the stop path.
		snapshot := recorder.Snapshot()
		cancel()
		<-done
		if len(snapshot) == 0 {
			fmt.Fprintln(stderr, "huh capture: no audio was captured")
			return 1
		}
		if err := exportWAV(*output, format, snapshot); err != nil {
			fmt.Fprintf(stderr, "huh capture: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Exported %s of recent audio to %s\n", durationForBytes(len(snapshot), format), *output)
		return 0
	}
}

func exportWAV(path string, format audio.Format, pcm []byte) error {
	// O_EXCL prevents a development capture from silently replacing an
	// existing recording or any unrelated user file.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	if err := audio.WriteWAV(file, format, pcm); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}
	return nil
}

func durationForBytes(size int, format audio.Format) time.Duration {
	bytesPerSecond, err := format.BytesPerSecond()
	if err != nil || bytesPerSecond == 0 {
		return 0
	}
	return time.Duration(size) * time.Second / time.Duration(bytesPerSecond)
}

func runDoctor(w io.Writer) int {
	fmt.Fprintf(w, "HUH? %s environment check\n", version)
	fmt.Fprintf(w, "  operating system: %s\n", runtime.GOOS)

	ready := true
	if runtime.GOOS != "linux" {
		fmt.Fprintln(w, "  linux target:     missing (the first release targets Linux)")
		ready = false
	} else {
		fmt.Fprintln(w, "  linux target:     ready")
	}

	ready = reportExecutable(w, "audio capture", "pw-cat") && ready
	ready = reportExecutable(w, "transcription", "whisper-cli") && ready

	if ready {
		fmt.Fprintln(w, "\nEnvironment is ready for the planned capture pipeline.")
		return 0
	}

	fmt.Fprintln(w, "\nCore development can continue, but live capture is not ready yet.")
	return 1
}

func reportExecutable(w io.Writer, label, name string) bool {
	path, err := exec.LookPath(name)
	if err != nil {
		fmt.Fprintf(w, "  %-17s missing (%s not found)\n", label+":", name)
		return false
	}
	fmt.Fprintf(w, "  %-17s ready (%s)\n", label+":", path)
	return true
}
