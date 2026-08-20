package audio

import (
	"errors"
	"testing"
	"time"
)

func TestSpeechPCMThirtySecondCapacity(t *testing.T) {
	capacity, err := SpeechPCM().BytesFor(30 * time.Second)
	if err != nil {
		t.Fatalf("BytesFor() error = %v", err)
	}
	if capacity != 960_000 {
		t.Fatalf("BytesFor() = %d, want 960000", capacity)
	}
}

func TestBytesForRoundsUpPartialSamples(t *testing.T) {
	format := Format{SampleRate: 1, Channels: 1, BytesPerSample: 1}

	capacity, err := format.BytesFor(1500 * time.Millisecond)
	if err != nil {
		t.Fatalf("BytesFor() error = %v", err)
	}
	if capacity != 2 {
		t.Fatalf("BytesFor() = %d, want 2", capacity)
	}
}

func TestInvalidFormat(t *testing.T) {
	_, err := (Format{}).BytesFor(time.Second)
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("BytesFor() error = %v, want ErrInvalidFormat", err)
	}
}
