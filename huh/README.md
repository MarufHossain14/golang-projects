# HUH?

> Instant replay for the sentence you just missed.

HUH? is a privacy-focused desktop application that keeps a short, temporary
audio buffer in memory. When you realize you missed something, press a hotkey:
HUH? retrieves the audio from the previous few seconds, replays the relevant
speech, transcribes it locally, and briefly displays the result as subtitles.

It is a DVR for the immediate past—not a traditional recorder that has to be
started before something happens.

## The idea

Someone says:

> Put the package behind the recycling bin.

Five seconds later, you realize you did not understand them. Press the HUH?
hotkey and see:

```text
12 seconds ago:
"Put the package behind the recycling bin."

[Replay] [Dismiss]
```

The first version will target Linux with PipeWire and will run entirely on the
local machine.

## How it behaves

HUH? starts only when the user explicitly enables listening:

```bash
huh listen
```

After that, it remains active and continuously holds only the newest section of
audio—30 seconds by default. The user does not need to press anything before
someone speaks.

When the hotkey is pressed, HUH? takes a snapshot of what is already in the
buffer. New audio continues entering the live buffer while the snapshot is
replayed and transcribed.

Listening can be stopped at any time:

```bash
huh stop
```

Stopping clears the in-memory buffer immediately. Starting automatically at
login may be supported later, but it must be explicitly enabled.

## Processing flow

```text
Microphone or selected system audio
                 |
                 v
          Audio capture stream
                 |
                 v
       30-second circular buffer
       (old audio is overwritten)
                 |
          global hotkey pressed
                 |
                 v
        copy a buffer snapshot
            /           \
           v             v
   detect and replay   transcribe locally
      recent speech          |
                             v
                    temporary subtitle
```

The ring buffer is intentionally small. At 16 kHz, mono, 16-bit PCM, a
30-second buffer requires approximately 960 KB of memory.

## What happens after the hotkey is pressed

1. Copy the current circular buffer into an isolated snapshot.
2. Continue collecting new audio in the original buffer.
3. Use voice-activity detection to locate recent speech in the snapshot.
4. Add a small amount of silence before and after the detected speech so replay
   does not sound abruptly clipped.
5. Begin playback immediately.
6. Send the same segment to a local speech-to-text model.
7. Show the resulting text in a temporary on-screen overlay.
8. Discard the snapshot after playback and transcription are complete.

Playback should start nearly instantly. Transcription may arrive a moment later
and should never delay replay.

## Planned commands

```bash
# Begin maintaining the rolling in-memory buffer.
huh listen

# Stop listening and clear buffered audio.
huh stop

# Replay the most recent detected speech.
huh replay

# Transcribe recent speech and display the text.
huh read

# Show whether HUH? is listening and which source it uses.
huh status
```

The desktop hotkey should perform the common `replay` and `read` workflow
without requiring a terminal.

## Audio sources

The MVP will support a microphone source. Later versions may use PipeWire to
capture:

- audio from a selected application;
- movies, games, or lectures;
- voice calls; or
- all system output.

Audio-source selection must always be explicit. Application-specific capture is
preferable to silently capturing every system sound.

## Privacy principles

Privacy is part of the core design, not an optional setting.

- The rolling buffer lives in memory, not in a permanent recording.
- Old audio is continuously overwritten.
- Stopping the application clears buffered audio immediately.
- Transcription happens locally; no audio is uploaded to a cloud service.
- Snapshots are discarded after use unless the user explicitly saves one.
- A persistent tray icon or microphone indicator shows when listening is active.
- Launch-at-login is opt-in.
- The interface clearly identifies the active audio source.

The application cannot solve the social and legal questions around recording
other people. Users are responsible for consent and applicable local laws,
especially when capturing calls or nearby conversations.

## Proposed architecture

```text
cmd/huh                 CLI entry point and lifecycle
internal/audio          PipeWire capture and audio playback
internal/buffer         Thread-safe circular PCM buffer
internal/speech         Voice-activity detection and segment selection
internal/transcribe     Long-running local Whisper integration
internal/hotkey         Global keyboard shortcut
internal/overlay        Temporary subtitle window
internal/config         Source, duration, hotkey, and privacy settings
```

The speech-to-text model should remain loaded in a background process so each
hotkey press does not incur model startup time. A small local Whisper model is a
reasonable starting point; replay remains useful even when transcription is
still processing or produces an imperfect result.

## Three-week development plan

### Week 1: prove instant replay

- Capture microphone audio through PipeWire.
- Store PCM frames in a thread-safe circular buffer.
- Implement configurable 10–60 second buffer lengths.
- Snapshot the buffer without stopping live capture.
- Replay a snapshot on demand.
- Add `listen`, `stop`, `replay`, and `status` commands.

At the end of week one, pressing a command should reliably replay the past 30
seconds. This is the technical foundation of the product.

### Week 2: find and understand speech

- Add voice-activity detection.
- Extract the most recent speech segment with sensible padding.
- Integrate a local Whisper transcription process.
- Keep the model warm between requests.
- Add `read` and combined replay/transcribe behavior.
- Handle silence, multiple speech segments, and noisy rooms gracefully.

### Week 3: make it feel like a product

- Add a configurable global hotkey.
- Build a small transparent subtitle overlay.
- Add a visible listening indicator or tray control.
- Improve device selection and error messages.
- Test long-running capture, memory use, and repeated hotkey presses.
- Package the Linux application and write installation instructions.

## MVP success criteria

The first release is complete when it can:

- run continuously without growing memory usage;
- retain only a configurable, short period of audio;
- replay audio that occurred before the hotkey was pressed;
- isolate the most recent speech instead of replaying all 30 seconds;
- transcribe locally without network access;
- clearly show when listening is active; and
- clear all temporary audio when stopped.

## Non-goals for the first release

- Cloud transcription or synchronization
- Permanent conversation archives
- Speaker identification
- Mobile applications
- Support for every operating system
- Perfect transcription in crowded environments
- Secret or invisible background recording

## Main technical risks

- PipeWire capture and device selection can vary between Linux environments.
- Global hotkeys behave differently across Wayland compositors.
- Voice-activity detection may confuse music, television, or background noise
  with speech.
- Local transcription latency depends on the selected model and hardware.
- Microphone capture is less useful when the user is wearing headphones; system
  audio capture will eventually be important.

These risks are why the project begins with a command-driven replay prototype.
Once retrieving past audio is reliable, transcription and UI can be developed
without obscuring problems in the core pipeline.

## Project status

This repository currently contains the product concept and implementation plan.
The first engineering milestone is a minimal Go program that captures audio into
a circular buffer and replays a snapshot on command.
