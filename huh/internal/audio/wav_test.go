package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWriteWAV(t *testing.T) {
	pcm := []byte{1, 2, 3, 4}
	var output bytes.Buffer

	if err := WriteWAV(&output, SpeechPCM(), pcm); err != nil {
		t.Fatalf("WriteWAV() error = %v", err)
	}

	wav := output.Bytes()
	if got, want := len(wav), 44+len(pcm); got != want {
		t.Fatalf("WAV length = %d, want %d", got, want)
	}
	if got := string(wav[:4]); got != "RIFF" {
		t.Fatalf("WAV signature = %q, want RIFF", got)
	}
	if got := binary.LittleEndian.Uint32(wav[24:28]); got != 16_000 {
		t.Fatalf("sample rate = %d, want 16000", got)
	}
	if got := binary.LittleEndian.Uint32(wav[40:44]); got != uint32(len(pcm)) {
		t.Fatalf("data size = %d, want %d", got, len(pcm))
	}
	if got := wav[44:]; !bytes.Equal(got, pcm) {
		t.Fatalf("PCM payload = %v, want %v", got, pcm)
	}
}

func TestWriteWAVRejectsUnalignedPCM(t *testing.T) {
	var output bytes.Buffer

	err := WriteWAV(&output, SpeechPCM(), []byte{1})

	if err == nil {
		t.Fatal("WriteWAV() error = nil, want sample-alignment error")
	}
}
