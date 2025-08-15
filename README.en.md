# WasmP2PChat 🛰️

A chat application that utilizes the WebRTC P2P DataChannel, implemented with Go WebAssembly, allowing you to match and converse with random strangers.

## Summary ✴️

- **Purpose** ：A demo to learn the basics of integrating Wasm and WebRTC
- **Features** ：
  - Random matchmaking for P2P connection
  - Text message sending and receiving
- **Intended audience** ：Engineers who want to test WebRTC and Wasm (intermediate level)

## Usage 🧑‍💻
### Requirements

- Go 1.24.x（Wasm enabled）
- Any modern browser (Chrome / Firefox / Safari)
- 3 terminal windows
- Recommended to run in a local network (customization required for production use)

## How chat works 🧐

1. Build the WebAssembly file
```bash
GOOS=js GOARCH=wasm go build -o main.wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
```

2. Terminal 1: Start static file server
```bash
python3 -m http.server 8080
```

3. Terminal 2: Start matchmaking server
```bash
go run cmd/matchmaking/main.go
```

4. Terminal 3: Start signaling server
```bash
go run cmd/signaling/main.go
```

5. Open in browser
- Open http://localhost:8080 in two separate tabs
- Press START in both tabs
- Enter a message and press SEND to exchange messages

## Stack 🛠️

- WebAssembly (Wasm)
  - Binary format for running languages like Go in the browser
  - High performance, type-safe, and supports Go’s concurrency
- WebRTC
  - Real-time communication technology for direct browser-to-browser connections
  - Supports audio, video, and data
- DataChannel
  - WebRTC’s data communication feature
  - Sends text or binary data directly
- P2P Communication
  - Direct device-to-device communication without a server
  - Reduces server load and minimizes latency

## Benefits and limitations 👩‍⚕️

```bash
[Browser A] --WebSocket--> [Matchmaking Server] <--WebSocket-- [Browser B]
       ↓ RoomID notification                ↓
[Browser A] --WebSocket--> [Signaling Server] <--WebSocket-- [Browser B]
       ↓ Exchange connection info (SDP, ICE candidates)
[Browser A] <==== WebRTC DataChannel (P2P) ====> [Browser B]

```

1. Matchmaking
    - Pairs up two users and issues a Room ID

2. Signaling
    - Exchanges SDP and ICE candidates required for WebRTC connection

3. P2P Communication
    - Once connected, browsers exchange data directly

## Troubleshooting ⛓️‍💥

| Symptom                   | Cause                          | Solution          |
| ------------------------- | ------------------------------ | ----------------- |
| Cannot connect            | Signaling server not running   | Check Terminal 3  |
| Cannot find a peer        | Matchmaking server not running | Check Terminal 2  |
| Not working outside local | NAT traversal failure          | Add a TURN server |


## Reference 🔖

- [https://webrtc.org/?hl=ja](https://webrtc.org/?hl=ja)
- [https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts](https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts)
- [https://github.com/OpenAyame/ayame](https://github.com/OpenAyame/ayame)