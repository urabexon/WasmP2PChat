# WasmP2PChat 🛰️

Go製のWebAssembly（Wasm）とWebRTCを利用し、ブラウザ間で直接（P2P）チャットができるデモアプリケーションです。  
マッチメイキングサーバで相手を探し、シグナリングサーバを通じて接続情報を交換した後は、ブラウザ同士が直接メッセージを送受信します。

## Summary ✴️

- **目的** ：Wasm と WebRTC の基本的な連携方法を学ぶためのデモ
- **機能** ：
  - ランダムに相手を探してP2P接続
  - テキストメッセージの送受信
- **想定対象者** ：WebRTCやWasmの動作確認(中級エンジニア対象)

## Usage 🧑‍💻
### 必要環境

- Go 1.24.x（Wasm対応）
- 任意のモダンブラウザ（Chrome / Firefox / Safari）
- ターミナル3つ使用
- ローカルネットワークで動作確認推奨(本番環境等にしたい方はカスタマイズしてください)

## How chat works 🧐

1. WebAssemblyファイルのビルド
```bash
GOOS=js GOARCH=wasm go build -o main.wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
```

2. ターミナル1: 静的ファイルサーバ起動
```bash
python3 -m http.server 8080
```

3. ターミナル2: マッチメイキングサーバ起動
```bash
go run cmd/matchmaking/main.go
```

4. ターミナル3: シグナリングサーバ起動
```bash
go run cmd/signaling/main.go
```

5. ブラウザで接続
- http://localhost:8080 を2つのタブで開く
- 両方でSTARTを押す
- メッセージを入力して SEND で送受信

## Stack 🛠️

- WebAssembly (Wasm)
  - ブラウザでGoなどの言語を実行できるバイナリ形式
  - 高速かつ型安全、Goの並行処理を活用可能
- WebRTC
  - ブラウザ同士を直接つなぐリアルタイム通信技術
  - 音声、映像、データ通信に対応
- DataChannel
  - WebRTCのデータ通信機能
  - テキストやバイナリデータを直接送受信
- P2P通信
  - サーバーを経由せず、端末同士が直接やり取り
  - サーバー負荷を軽減し、遅延を最小化

## Benefits and limitations 👩‍⚕️

```bash
[Browser A] --WebSocket--> [Matchmaking Server] <--WebSocket-- [Browser B]
       ↓ RoomID通知                             ↓
[Browser A] --WebSocket--> [Signaling Server] <--WebSocket-- [Browser B]
       ↓ 接続情報交換（SDP, ICE候補）
[Browser A] <==== WebRTC DataChannel (P2P) ====> [Browser B]
```

1. マッチメイキング
    - 接続してきたユーザー同士をペアにしてRoom IDを発行

2. シグナリング
    - WebRTC接続に必要なSDPやICE候補を交換

3. P2P通信
    - 接続確立後はブラウザ間で直接データ送受信

## Troubleshooting ⛓️‍💥

| 症状          | 原因             | 対処        |
| ----------- | -------------- | --------- |
| 接続できない      | シグナリングサーバ未起動   | ターミナル3を確認 |
| 相手が見つからない   | マッチメイキングサーバ未起動 | ターミナル2を確認 |
| ローカル以外で動かない | NAT越え不可        | TURNサーバ追加 |

## Reference 🔖

- [https://webrtc.org/?hl=ja](https://webrtc.org/?hl=ja)
- [https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts](https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts)
- [https://github.com/OpenAyame/ayame](https://github.com/OpenAyame/ayame)