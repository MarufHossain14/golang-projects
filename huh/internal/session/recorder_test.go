package session

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/MarufHossain14/golang-projects/huh/internal/buffer"
)

func TestRecorderCapturesAndClearsCompletedSource(t *testing.T) {
	ring, _ := buffer.New(4)
	recorder, _ := NewRecorder(ring)
	source := io.NopCloser(bytes.NewBufferString("abcdef"))

	if err := recorder.Run(context.Background(), source); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if recorder.Running() {
		t.Fatal("Running() = true after source completed")
	}
	if got := recorder.Snapshot(); len(got) != 0 {
		t.Fatalf("Snapshot() = %q after stop, want cleared", got)
	}
}

func TestRecorderExposesSnapshotWhileRunning(t *testing.T) {
	ring, _ := buffer.New(8)
	recorder, _ := NewRecorder(ring)
	source := newBlockingSource([]byte("speech"))
	done := make(chan error, 1)

	go func() {
		done <- recorder.Run(context.Background(), source)
	}()

	select {
	case <-source.read:
	case <-time.After(time.Second):
		t.Fatal("recorder did not read source")
	}

	if got, want := recorder.Snapshot(), []byte("speech"); !bytes.Equal(got, want) {
		t.Fatalf("Snapshot() = %q, want %q", got, want)
	}
	_ = source.Close()
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRecorderRejectsConcurrentRun(t *testing.T) {
	ring, _ := buffer.New(8)
	recorder, _ := NewRecorder(ring)
	source := newBlockingSource(nil)
	done := make(chan error, 1)
	go func() { done <- recorder.Run(context.Background(), source) }()

	select {
	case <-source.read:
	case <-time.After(time.Second):
		t.Fatal("recorder did not start")
	}

	err := recorder.Run(context.Background(), io.NopCloser(bytes.NewReader(nil)))
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("Run() error = %v, want ErrAlreadyRunning", err)
	}
	_ = source.Close()
	<-done
}

type blockingSource struct {
	mu        sync.Mutex
	initial   []byte
	read      chan struct{}
	closed    chan struct{}
	readOnce  sync.Once
	closeOnce sync.Once
}

func newBlockingSource(initial []byte) *blockingSource {
	return &blockingSource{
		initial: initial,
		read:    make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

func (s *blockingSource) Read(p []byte) (int, error) {
	s.mu.Lock()
	if s.initial != nil {
		n := copy(p, s.initial)
		s.initial = nil
		s.readOnce.Do(func() { close(s.read) })
		s.mu.Unlock()
		return n, nil
	}
	s.readOnce.Do(func() { close(s.read) })
	s.mu.Unlock()
	<-s.closed
	return 0, io.EOF
}

func (s *blockingSource) Close() error {
	s.readOnce.Do(func() { close(s.read) })
	s.closeOnce.Do(func() { close(s.closed) })
	return nil
}
