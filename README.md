# Pi Audio Recorder Client
**Version:** 1.0.0
- Initial stable release

Small recorder client that connects to a backend via WebSocket and records audio on the Pi.  
This document describes architecture, WebSocket commands, payloads, how to run, and recommended usage.

## Development Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure
3. Install Go dependencies: `go mod download`
4. (On Raspberry Pi only) Build RNNoise:
   ```bash
   cd rnnoise
   ./autogen.sh
   ./configure
   make
   ```
5. Install sox: `sudo apt-get install sox` (Raspberry Pi only)

## Important Files

- `.env` - Local configuration (NOT committed to Git)
- `recordings/` - Audio files (NOT committed to Git)
- `rnnoise/.libs/` - Compiled binaries (NOT committed to Git, build on target)


## Quick links (open these in your workspace)
- [cmd/main.go](cmd/main.go) — program entry

- [internal/wsclient/client.go](internal/wsclient/client.go) — WS client connect/read/write and [`wsclient.Start`](internal/wsclient/client.go)

- [internal/wsclient/messages.go](internal/wsclient/messages.go) — message payload types including [`wsclient.StartRecordingMessage`](internal/wsclient/messages.go)

- [internal/wsclient/handlers.go](internal/wsclient/handlers.go) — message handlers incl. [`wsclient.handleStartRecordingMulti`](internal/wsclient/handlers.go)

- [internal/recorder/multi_recorder.go](internal/recorder/multi_recorder.go) — session lifecycle and [`recorder.StartSession`](internal/recorder/multi_recorder.go) / [`recorder.StopSession`](internal/recorder/multi_recorder.go)

- [internal/recorder/session_manager.go](internal/recorder/session_manager.go)

- [internal/recorder/session.go](internal/recorder/session.go)

- [internal/audio/utils.go](internal/audio/utils.go) — device helpers such as [`audio.GetDeviceIndexByName`](internal/audio/utils.go) and [`audio.ListAudioDevices`](internal/audio/utils.go)

- [internal/audio/aiff.go](internal/audio/aiff.go) — AIFF implementation and recording loop

- [internal/config/config.go](internal/config/config.go) — environment-driven config

- [test_devices.go](test_devices.go) — quick device listing helper

- [test_recording.go](test_recording.go) — simple recording test






## Architecture overview
- WS client (`internal/wsclient`) connects to backend at `ws://<BACKEND_HOST><WebSocketPath>?pi_id=<PI_ID>` — see [`wsclient.connect`](internal/wsclient/client.go).
- Backend sends commands over WS. Handlers in [`internal/wsclient/handlers.go`](internal/wsclient/handlers.go) parse and call recorder functions.
- Recorder (`internal/recorder`) manages sessions with [`recorder.StartSession`](internal/recorder/multi_recorder.go) and [`recorder.StopSession`](internal/recorder/multi_recorder.go).
- Audio code in `internal/audio` uses PortAudio via `github.com/gordonklaus/portaudio`; devices are enumerated in [`audio.ListAudioDevices`](internal/audio/utils.go).






## WebSocket commands & payloads
Top-level envelope (text JSON):
{
  "type": "<command_type>",
  "data": { ... command payload ... }
}

Targeting / pi_id
- Clients include their pi_id as a query parameter when connecting: ws://host/ws?pi_id=<pi_id>
- Backend should route messages to the intended PI by sending to that has the pi_id.
- Alternatively the backend can include target_pi_id in the message JSON; the Pi will ignore messages whose target_pi_id does not match its own id.
- Example envelope targeting a specific Pi:
+ ```json
+ {
+   "type": "start_recording",
+   "target_pi_id": "pi-0123456789abcdef",
+   "data": {
+     "command": "start_recording",
+     "session_id": "mic1",
+     "device_name": "USB Condenser Microphone"
+   }
+ }

Supported types:
- `start_recording` — start a session
  - Payload shape: [`wsclient.StartRecordingMessage`](internal/wsclient/messages.go)
    - `session_id` (string) — required
    - `device_index` (int) — optional if `device_name` provided
    - `device_name` (string) — optional; Pi resolves to index using [`audio.GetDeviceIndexByName`](internal/audio/utils.go)
  - Example (by name):
    {
      "type":"start_recording",
      "data":{"command":"start_recording","session_id":"mic1","device_name":"usb condenser"}
    }
  - Example (by index):
    {
      "type":"start_recording",
      "data":{"command":"start_recording","session_id":"mic1","device_index":0}
    }

- `stop_recording` — stop a session
  - Payload shape: [`wsclient.StopRecordingMessage`](internal/wsclient/messages.go)
    - `session_id` (string) — required

- `list_devices` — request device list  
  - Response: the Pi returns the device list in JSON (easy for the backend to parse). Example response:
  ```json
  [
    {
      "index": 0,
      "name": "USB Condenser Microphone: Audio (hw:2,0)",
      "max_input_channels": 1,
      "default_sample_rate": 44100,
      "host_api": "ALSA"
    }
  ]
  ```

- `stop_all` — stop all active sessions (`recorder.StopAllSessions`)

Handlers that process these are in [`internal/wsclient/handlers.go`](internal/wsclient/handlers.go), e.g. [`wsclient.handleStartRecordingMulti`](internal/wsclient/handlers.go) resolves device name (if present) before calling [`recorder.StartSession`](internal/recorder/multi_recorder.go).






## How Start/Stop works internally
- Start:
  - Backend message → [`wsclient.handleStartRecordingMulti`](internal/wsclient/handlers.go) → resolves device name to index (if needed) → calls [`recorder.StartSession`](internal/recorder/multi_recorder.go).

  - [`recorder.StartSession`](internal/recorder/multi_recorder.go) creates session via [`session_manager.CreateSession`](internal/recorder/session_manager.go), instantiates audio instance via `audio.NewAudioInstance(...)`, sets device index and initializes the recorder, then starts recording goroutine (`IAudioFormat.Record` in [internal/audio/aiff.go](internal/audio/aiff.go)).
- Stop:
  - Backend message → [`wsclient.handleStopRecordingSession`](internal/wsclient/handlers.go) → calls [`recorder.StopSession`](internal/recorder/multi_recorder.go) → sends stop control via recorder control channel and waits for confirmation → session removed and file path returned. The handler begins an upload via `sendFile`.






## Running locally / testing
- Build/run client:
  - Run directly: `go run cmd/main.go`
  - Or build: `go build -o pi-client ./cmd && ./pi-client`
- Configure Pi ID (env): `PI_ID` (defaults to `pi01`) — see [cmd/main.go](cmd/main.go).
- Configurable environment variables (defaults inside [`internal/config/config.go`](internal/config/config.go)):
  - `SYS_RECORD_PATH` (default `./recordings`)
  - `SYS_AUDIO_TYPE` (0 = aiff (default), 1 = wav)
  - `SYS_AUDIO_CHANNEL`
  - `SYS_AUDIO_SAMPLE_RATE`
  - `SYS_AUDIO_INPUT_BUFFER_SIZE`

- Quick device listing:
  - Run: `go run test_devices.go` — uses [`audio.PrintAvailableDevices`](internal/audio/utils.go).

- Quick recording test:
  - See [test_recording.go](test_recording.go) which resolves by name via [`audio.GetDeviceIndexByName`](internal/audio/utils.go) and uses [`recorder.StartSession`](internal/recorder/multi_recorder.go).






## File locations for produced recordings
Recordings are written under `SYS_RECORD_PATH` (default `./recordings`) with per-session directories; Example: `recordings/mic1/device_0_20251203_160611.aiff`.





## Example backend snippet (Node.js)
```js
const WebSocket = require('ws');
const ws = new WebSocket('ws://aeronsarondo.site/ws?pi_id=pi01');

ws.on('open', () => {
  const msg = {
    type: 'start_recording',
    data: {
      command: 'start_recording',
      session_id: 'mic1',
      device_name: 'usb condenser'
    }
  };
  ws.send(JSON.stringify(msg));
});