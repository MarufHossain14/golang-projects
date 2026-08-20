// Package session coordinates an audio source with the rolling memory buffer.
package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/MarufHossain14/golang-projects/huh/internal/buffer"
)

const defaultReadSize = 4096

var ErrAlreadyRunning = errors.New("recording session is already running")

// Recorder owns the lifecycle of one live source. Source adapters normalize
// their audio before passing it here, leaving Recorder independent of PipeWire.
type Recorder struct {
	buffer *buffer.Ring

	mu      sync.Mutex
	running bool
}

func NewRecorder(ring *buffer.Ring) (*Recorder, error) {
	if ring == nil {
		return nil, fmt.Errorf("audio buffer is required")
	}
	return &Recorder{buffer: ring}, nil
}

// Run copies bytes from source until it reaches EOF, the context is cancelled,
// or reading fails. Cancelling closes the source so a blocked device read can
// return; ownership of source therefore transfers to Run for its duration.
func (r *Recorder) Run(ctx context.Context, source io.ReadCloser) error {
	if source == nil {
		return fmt.Errorf("audio source is required")
	}
	if err := r.begin(); err != nil {
		return err
	}
	defer r.end()
	defer source.Close()
	defer r.buffer.Reset()

	closed := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = source.Close()
		case <-closed:
		}
	}()
	defer close(closed)

	chunk := make([]byte, defaultReadSize)
	for {
		n, err := source.Read(chunk)
		if n > 0 {
			_, _ = r.buffer.Write(chunk[:n])
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("read audio source: %w", err)
		}
	}
}

func (r *Recorder) Snapshot() []byte {
	return r.buffer.Snapshot()
}

func (r *Recorder) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func (r *Recorder) begin() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		return ErrAlreadyRunning
	}
	r.running = true
	return nil
}

func (r *Recorder) end() {
	r.mu.Lock()
	r.running = false
	r.mu.Unlock()
}
