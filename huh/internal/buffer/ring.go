// Package buffer provides fixed-size in-memory storage for recent PCM audio.
package buffer

import (
	"fmt"
	"sync"
)

// Ring retains only the newest capacity bytes written to it. Ring is safe for
// one capture goroutine to write while command handlers take snapshots.
type Ring struct {
	mu    sync.RWMutex
	data  []byte
	start int
	len   int
}

func New(capacity int) (*Ring, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("buffer capacity must be positive")
	}
	return &Ring{data: make([]byte, capacity)}, nil
}

func (r *Ring) Capacity() int {
	return len(r.data)
}

func (r *Ring) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.len
}

// Write appends p and discards the oldest bytes when capacity is exceeded. A
// write larger than the whole ring intentionally keeps only its final bytes.
func (r *Ring) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	written := len(p)
	if len(p) >= len(r.data) {
		p = p[len(p)-len(r.data):]
		copy(r.data, p)
		r.start = 0
		r.len = len(r.data)
		return written, nil
	}

	for len(p) > 0 {
		end := (r.start + r.len) % len(r.data)
		chunk := min(len(p), len(r.data)-end)
		copy(r.data[end:end+chunk], p[:chunk])
		p = p[chunk:]

		free := len(r.data) - r.len
		if chunk <= free {
			r.len += chunk
			continue
		}

		overwritten := chunk - free
		r.start = (r.start + overwritten) % len(r.data)
		r.len = len(r.data)
	}

	return written, nil
}

// Snapshot returns an independent chronological copy, from the oldest retained
// byte to the newest. Callers may safely modify or discard it after the lock is
// released while capture continues.
func (r *Ring) Snapshot() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make([]byte, r.len)
	if r.len == 0 {
		return snapshot
	}

	first := min(r.len, len(r.data)-r.start)
	copy(snapshot, r.data[r.start:r.start+first])
	copy(snapshot[first:], r.data[:r.len-first])
	return snapshot
}

// Reset clears logical contents and zeroes the backing memory so stopping a
// session does not leave recoverable audio in the allocated ring.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	clear(r.data)
	r.start = 0
	r.len = 0
}
