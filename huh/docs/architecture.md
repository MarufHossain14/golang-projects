# Architecture Notes

This document expands the design constraints for HUH?'s first implementation.
The exact libraries can be selected during the initial technical spike.

## Runtime model

HUH? is best implemented as one long-running user process. The CLI communicates
with that process for commands such as `replay`, `read`, `status`, and `stop`.
This avoids restarting audio capture or loading the transcription model for each
request.

```text
                           local command channel
CLI / hotkey / tray  ------------------------------+
                                                     |
                                                     v
+-----------------------------------------------------------+
|                      HUH? daemon                           |
|                                                           |
|  capture -> ring buffer -> snapshot -> speech detection   |
|                                      |             |      |
|                                      v             v      |
|                                   playback     transcription|
|                                                     |     |
|                                                     v     |
|                                                  overlay  |
+-----------------------------------------------------------+
```

## Audio representation

Normalize captured audio early to a single internal representation:

- signed 16-bit PCM;
- 16 kHz sample rate; and
- one channel.

This format is compact, suitable for speech, and accepted by common local
speech-to-text pipelines. Capture devices may provide a different format, so
resampling and channel conversion should happen before frames enter the ring
buffer.

## Circular buffer

The circular buffer has a fixed capacity derived from sample rate, sample size,
channel count, and configured duration. New frames overwrite the oldest frames.

A hotkey press must create a consistent snapshot without pausing capture long
enough to lose audio. The first implementation can use a short lock while
copying roughly one megabyte. More complicated lock-free designs should only be
considered if measurement shows the simple approach causes dropped frames.

The live buffer and captured snapshots have different lifetimes:

- the live buffer exists only while listening;
- a snapshot exists only while a replay/transcription request is active; and
- neither is persisted by default.

## Speech selection

Voice-activity detection operates on small frames from the snapshot. The segment
selector should work backward from the newest audio and find the most recent
contiguous speech region. Short silent gaps inside an utterance should be joined,
while long gaps should separate utterances.

Useful initial rules:

- add padding before and after speech;
- ignore extremely short bursts;
- limit the maximum replay duration; and
- fall back to the complete buffer when detection confidence is too low and the
  user explicitly requests raw replay.

## Concurrency

The application has several independent activities:

- reading frames from PipeWire;
- writing frames into the ring buffer;
- receiving CLI or hotkey requests;
- playing snapshots;
- transcribing snapshots; and
- updating the overlay.

Go routines and channels are appropriate for lifecycle events and bounded work
queues. Audio callbacks should do minimal work and should not block on playback,
transcription, or UI operations.

Only a small number of snapshots should be processed concurrently. Repeated
hotkey presses can cancel an older request, replace it, or enter a bounded queue;
they must not create unlimited transcription jobs.

## Failure behavior

Replay is the primary capability and should remain available if transcription
fails. Likewise, a missing overlay should not prevent CLI output.

Expected degraded states include:

- no usable capture device;
- selected device disappears;
- PipeWire restarts;
- transcription model is missing;
- transcription exceeds its deadline;
- no speech is detected; and
- global hotkey registration is unavailable.

Each state should produce a clear status message and preserve whatever subset of
the application remains useful.

## Security and privacy boundaries

The local command channel should be accessible only to the current user. Debug
logging must never include raw audio, transcripts, or buffer contents by
default. Crash reports must not attach audio snapshots.

If saving clips is added later, it should be a separate, explicit command with a
visible destination—not a side effect of replay or transcription.
