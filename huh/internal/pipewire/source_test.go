package pipewire

import (
	"slices"
	"testing"
)

func TestArguments(t *testing.T) {
	args := arguments("alsa_input.usb-mic")

	for _, want := range []string{
		"--record",
		"--raw",
		"--rate=16000",
		"--channels=1",
		"--channel-map=mono",
		"--format=s16",
		"--target=alsa_input.usb-mic",
		"-",
	} {
		if !slices.Contains(args, want) {
			t.Errorf("arguments() = %q, missing %q", args, want)
		}
	}
}

func TestArgumentsOmitsEmptyTarget(t *testing.T) {
	args := arguments("")
	if slices.Contains(args, "--target=") {
		t.Fatalf("arguments() = %q, want no empty target", args)
	}
}
