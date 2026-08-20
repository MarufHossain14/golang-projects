package buffer

import (
	"bytes"
	"sync"
	"testing"
)

func TestRingRetainsNewestBytes(t *testing.T) {
	ring, err := New(5)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _ = ring.Write([]byte("abc"))
	_, _ = ring.Write([]byte("def"))

	if got, want := ring.Snapshot(), []byte("bcdef"); !bytes.Equal(got, want) {
		t.Fatalf("Snapshot() = %q, want %q", got, want)
	}
}

func TestRingLargeWriteRetainsTail(t *testing.T) {
	ring, _ := New(4)

	n, err := ring.Write([]byte("abcdefgh"))

	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 8 {
		t.Fatalf("Write() = %d, want 8", n)
	}
	if got, want := ring.Snapshot(), []byte("efgh"); !bytes.Equal(got, want) {
		t.Fatalf("Snapshot() = %q, want %q", got, want)
	}
}

func TestSnapshotIsIndependent(t *testing.T) {
	ring, _ := New(4)
	_, _ = ring.Write([]byte("abcd"))

	snapshot := ring.Snapshot()
	snapshot[0] = 'z'

	if got, want := ring.Snapshot(), []byte("abcd"); !bytes.Equal(got, want) {
		t.Fatalf("Snapshot() = %q, want %q", got, want)
	}
}

func TestResetClearsRing(t *testing.T) {
	ring, _ := New(4)
	_, _ = ring.Write([]byte("abcd"))

	ring.Reset()

	if ring.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", ring.Len())
	}
	if got := ring.Snapshot(); len(got) != 0 {
		t.Fatalf("Snapshot() = %q, want empty", got)
	}
}

func TestConcurrentWriteAndSnapshot(t *testing.T) {
	ring, _ := New(128)
	var writers sync.WaitGroup

	for i := 0; i < 4; i++ {
		writers.Add(1)
		go func(value byte) {
			defer writers.Done()
			for j := 0; j < 1_000; j++ {
				_, _ = ring.Write(bytes.Repeat([]byte{value}, 8))
				_ = ring.Snapshot()
			}
		}(byte(i))
	}

	writers.Wait()
	if got := ring.Len(); got != ring.Capacity() {
		t.Fatalf("Len() = %d, want %d", got, ring.Capacity())
	}
}
