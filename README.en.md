# WasmP2PChat 🛰️

A chat application that utilizes the WebRTC P2P DataChannel, implemented with Go WebAssembly, allowing you to match and converse with random strangers.

## build

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
# 初回のみ
curl -O https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js
# サーバー起動
python3 -m http.server 8080
# アクセス
open http://localhost:8080



cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
GOOS=js GOARCH=wasm go build -o main.wasm
python3 -m http.server 8080

```