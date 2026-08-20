package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// WriteWAV wraps normalized PCM bytes in a standard little-endian WAV
// container. Explicit export is kept separate from the rolling buffer so the
// normal listening path never writes audio to disk by accident.
func WriteWAV(w io.Writer, format Format, pcm []byte) error {
	if err := format.Validate(); err != nil {
		return err
	}
	if format.BytesPerSample != 2 {
		return fmt.Errorf("WAV export currently supports 16-bit PCM only")
	}
	if len(pcm)%format.BytesPerSample != 0 {
		return fmt.Errorf("PCM length %d is not sample-aligned", len(pcm))
	}
	if len(pcm) > math.MaxUint32-36 {
		return fmt.Errorf("PCM data is too large for a WAV file")
	}

	byteRate, _ := format.BytesPerSecond()
	blockAlign := format.Channels * format.BytesPerSample
	header := wavHeader{
		RIFF:          [4]byte{'R', 'I', 'F', 'F'},
		FileSize:      uint32(36 + len(pcm)),
		WAVE:          [4]byte{'W', 'A', 'V', 'E'},
		FMT:           [4]byte{'f', 'm', 't', ' '},
		FormatSize:    16,
		AudioFormat:   1,
		Channels:      uint16(format.Channels),
		SampleRate:    uint32(format.SampleRate),
		ByteRate:      uint32(byteRate),
		BlockAlign:    uint16(blockAlign),
		BitsPerSample: uint16(format.BytesPerSample * 8),
		Data:          [4]byte{'d', 'a', 't', 'a'},
		DataSize:      uint32(len(pcm)),
	}

	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("write WAV header: %w", err)
	}
	if _, err := w.Write(pcm); err != nil {
		return fmt.Errorf("write WAV audio: %w", err)
	}
	return nil
}

type wavHeader struct {
	RIFF          [4]byte
	FileSize      uint32
	WAVE          [4]byte
	FMT           [4]byte
	FormatSize    uint32
	AudioFormat   uint16
	Channels      uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Data          [4]byte
	DataSize      uint32
}
