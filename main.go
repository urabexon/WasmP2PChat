//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"syscall/js"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/urabexon/WasmP2PChat/go-ayame"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	wsScheme          string
	matchmakingOrigin string
	signalingOrigin   string
)

type mmReqMsg struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type mmResMsg struct {
	Type      string    `json:"type"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func init() {
	loc := js.Global().Get("location")
	if loc.Truthy() {
		proto := loc.Get("protocol").String() // "http:" or "https:"
        if proto == "https:" {
            wsScheme = "wss"
        } else {
            wsScheme = "ws"
        }
		host := loc.Get("host").String()
		// JSから上書き可能: window.MATCHMAKING_ORIGIN / window.SIGNALING_ORIGIN
        if v := js.Global().Get("MATCHMAKING_ORIGIN"); v.Truthy() {
            matchmakingOrigin = v.String()
        } else {
            matchmakingOrigin = host
        }
        if v := js.Global().Get("SIGNALING_ORIGIN"); v.Truthy() {
            signalingOrigin = v.String()
        } else {
            signalingOrigin = host
        }
	}
}

func shortHash(now time.Time) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(now.String())); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:7], nil
}

func onMessage() func(webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			logElem(fmt.Sprintf("[Any]: %s\n", msg.Data))
		}
	}
}

func logElem(msg string) {
    el := getElementByID("logs")
    // textarea は value を更新しないと表示されません
    el.Set("value", el.Get("value").String()+msg)
    // 自動スクロール
    el.Set("scrollTop", el.Get("scrollHeight"))
}

func handleError() {
	logElem("[Sys]: Maybe Any left, Please restart\n")
}

func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

func main() {
	mmURL := url.URL{Scheme: wsScheme, Host: matchmakingOrigin, Path: "/matchmaking"}
	signalingURL := url.URL{Scheme: wsScheme, Host: signalingOrigin, Path: "/signaling"}

	now := time.Now()
	userID, _ := shortHash(now)
	reqMsg, err := json.Marshal(mmReqMsg{
		UserID:    userID,
		CreatedAt: now,
	})
	if err != nil {
		log.Fatal(err)
	}
	var resMsg mmResMsg
	var dc *webrtc.DataChannel
	defer func() {
		if dc != nil {
			dc.Close()
		}
	}()

	var conn *ayame.Connection
	connected := make(chan bool)
	js.Global().Set("startNewChat", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			ws, _, err := websocket.Dial(context.Background(), mmURL.String(), nil)
            if err != nil {
                logElem(fmt.Sprintf("[Err]: matchmaking dial failed: %v\n", err))
                return
            }
			defer ws.Close(websocket.StatusNormalClosure, "close connection")
			if err := ws.Write(context.Background(), websocket.MessageText, reqMsg); err != nil {
				logElem(fmt.Sprintf("[Err]: matchmaking write failed: %v\n", err))
			    return
	        }
			logElem("[Sys]: Waiting match...\n")
			for {
				if err := wsjson.Read(context.Background(), ws, &resMsg); err != nil {
					logElem(fmt.Sprintf("[Err]: matchmaking read failed: %v\n", err))
                    break
	            }
				logElem(fmt.Sprintf("[DBG]: received type=%s room=%s from matchmaking\n", resMsg.Type, resMsg.RoomID))
				if resMsg.Type == "MATCH" {
					break
				}
			}
			ws.Close(websocket.StatusNormalClosure, "close connection")
			if resMsg.Type == "MATCH" && signalingOrigin != "" {
				conn = ayame.NewConnection(signalingURL.String(), resMsg.RoomID, ayame.DefaultOptions(), false, false)
				conn.OnOpen(func(metadata *interface{}) {
					log.Println("Open")
					var err error
					dc, err = conn.CreateDataChannel("matchmaking-example", nil)
					if err != nil && err != fmt.Errorf("client does not exist") {
						log.Printf("CreateDataChannel error: %v", err)
						return
					}
					log.Printf("CreateDataChannel: label=%s", dc.Label())
					dc.OnMessage(onMessage())
					dc.OnOpen(func() {
				        logElem("[Sys]: DataChannel OPEN — 送受信できます\n")
				    })
				})
				conn.OnConnect(func() {
					logElem("[Sys]: Matching! (signaling open) — ICE候補交換中...\n")
					connected <- true
				})
				conn.OnDataChannel(func(c *webrtc.DataChannel) {
					log.Printf("OnDataChannel: label=%s", c.Label())
					if dc == nil {
						dc = c
					}
					dc.OnMessage(onMessage())
					dc.OnOpen(func() {
				        logElem("[Sys]: DataChannel OPEN (remote) — 送受信できます\n")
				    })
				})
				if err := conn.Connect(); err != nil {
				    logElem(fmt.Sprintf("[Err]: Failed to connect Ayame: %v\n", err))
				    return
				}
				select {
				case <-connected:
					return
				}
			}
		}()
		return js.Undefined()
	}))
	js.Global().Set("sendMessage", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			el := getElementByID("message")
			message := el.Get("value").String()
			if message == "" {
				js.Global().Call("alert", "Message must not be empty")
				return
			}
			if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
			    js.Global().Call("alert", "未接続です。まず START を押してマッチングしてください。")
			    return
            }
			if err := dc.SendText(message); err != nil {
				handleError()
				return
			}
			logElem(fmt.Sprintf("[You]: %s\n", message))
			el.Set("value", "")
		}()
		return js.Undefined()
	}))
	select {}
}
