package audio

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidFormat = errors.New("invalid audio format")

// Format describes the normalized PCM representation shared by capture,
// buffering, speech detection, and transcription. Keeping one internal format
// prevents every downstream component from needing its own resampling logic.
type Format struct {
	SampleRate     int
	Channels       int
	BytesPerSample int
}

func SpeechPCM() Format {
	return Format{
		SampleRate:     16_000,
		Channels:       1,
		BytesPerSample: 2,
	}
}

func (f Format) Validate() error {
	if f.SampleRate <= 0 {
		return fmt.Errorf("%w: sample rate must be positive", ErrInvalidFormat)
	}
	if f.Channels <= 0 {
		return fmt.Errorf("%w: channel count must be positive", ErrInvalidFormat)
	}
	if f.BytesPerSample <= 0 {
		return fmt.Errorf("%w: bytes per sample must be positive", ErrInvalidFormat)
	}
	return nil
}

func (f Format) BytesPerSecond() (int, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}
	return f.SampleRate * f.Channels * f.BytesPerSample, nil
}

func (f Format) BytesFor(duration time.Duration) (int, error) {
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	bytesPerSecond, err := f.BytesPerSecond()
	if err != nil {
		return 0, err
	}
	return int((int64(bytesPerSecond)*duration.Nanoseconds() + int64(time.Second) - 1) / int64(time.Second)), nil
}
