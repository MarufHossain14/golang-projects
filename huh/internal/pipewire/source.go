// Package pipewire adapts the official pw-cat utility to HUH?'s audio source.
package pipewire

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Source streams 16 kHz, mono, signed 16-bit PCM from a PipeWire capture node.
type Source struct {
	command *exec.Cmd
	stdout  io.ReadCloser
	stderr  bytes.Buffer

	waitOnce sync.Once
	waitErr  error
}

func Open(ctx context.Context, target string) (*Source, error) {
	path, err := exec.LookPath("pw-cat")
	if err != nil {
		return nil, fmt.Errorf("find pw-cat: %w", err)
	}

	command := exec.CommandContext(ctx, path, arguments(target)...)
	source := &Source{command: command}
	command.Stderr = &source.stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("open pw-cat output: %w", err)
	}
	source.stdout = stdout
	if err := command.Start(); err != nil {
		_ = stdout.Close()
		return nil, fmt.Errorf("start pw-cat: %w", err)
	}
	return source, nil
}

func arguments(target string) []string {
	args := []string{
		"--record",
		"--raw",
		"--rate=16000",
		"--channels=1",
		"--channel-map=mono",
		"--format=s16",
	}
	if target != "" {
		args = append(args, "--target="+target)
	}
	return append(args, "-")
}

func (s *Source) Read(p []byte) (int, error) {
	n, err := s.stdout.Read(p)
	if !errors.Is(err, io.EOF) {
		return n, err
	}

	if waitErr := s.wait(); waitErr != nil {
		message := s.stderr.String()
		if message != "" {
			return n, fmt.Errorf("pw-cat stopped: %w: %s", waitErr, message)
		}
		return n, fmt.Errorf("pw-cat stopped: %w", waitErr)
	}
	return n, io.EOF
}

func (s *Source) Close() error {
	_ = s.stdout.Close()
	if s.command.Process != nil {
		_ = s.command.Process.Kill()
	}
	return s.wait()
}

func (s *Source) wait() error {
	s.waitOnce.Do(func() {
		s.waitErr = s.command.Wait()
	})
	return s.waitErr
}
