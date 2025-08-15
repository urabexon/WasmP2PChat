package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type peer struct {
	c *websocket.Conn
}

type room struct {
	a *peer
	b *peer
}

var (
	mu    sync.Mutex
	rooms = map[string]*room{}
)

func acceptWS(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // 開発中は緩め、運用は限定
	})
}

func writeJSON(ctx context.Context, c *websocket.Conn, v any) error {
	return wsjson.Write(ctx, c, v)
}

func forwardRaw(ctx context.Context, to *websocket.Conn, raw map[string]json.RawMessage) error {
	// 受け取ったキーをそのまま中継（ice など未知キーも保持）
	return writeJSON(ctx, to, raw)
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := acceptWS(w, r)
	if err != nil {
		log.Println("accept err:", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "server error")

	ctx := r.Context()

	// 1) register を「生JSON」で読む
	var reg map[string]json.RawMessage
	if err := wsjson.Read(ctx, c, &reg); err != nil {
		log.Println("read register err:", err)
		return
	}
	var typ string
	_ = json.Unmarshal(reg["type"], &typ)
	if typ != "register" {
		log.Println("invalid first message, want register")
		return
	}
	var roomID string
	_ = json.Unmarshal(reg["roomId"], &roomID)
	if roomID == "" {
		log.Println("register missing roomId")
		return
	}

	// 2) 部屋に入れる & accept を返す
	mu.Lock()
	rm := rooms[roomID]
	if rm == nil {
		rm = &room{}
		rooms[roomID] = rm
	}
	exist := false
	if rm.a == nil {
		rm.a = &peer{c: c}
	} else if rm.b == nil {
		rm.b = &peer{c: c}
		exist = true
	} else {
		mu.Unlock()
		log.Println("room full:", roomID)
		c.Close(websocket.StatusPolicyViolation, "room full")
		return
	}
	mu.Unlock()

	acc := map[string]any{
		"type": "accept",
		"iceServers": []map[string]any{
			{"urls": []string{"stun:stun.l.google.com:19302"}},
		},
		"isExistClient": exist,
	}
	if err := writeJSON(ctx, c, acc); err != nil {
		log.Println("write accept err:", err)
		return
	}
	log.Printf("registered room=%s exist=%v\n", roomID, exist)

	// 3) 受信ループ：offer/answer/candidate/ping などを「生のまま」相手へ転送
	for {
		var msg map[string]json.RawMessage
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			log.Println("read err:", err)
			break
		}
		var mt string
		_ = json.Unmarshal(msg["type"], &mt)

		if mt == "ping" {
			_ = writeJSON(ctx, c, map[string]string{"type": "pong"})
			continue
		}

		mu.Lock()
		rm = rooms[roomID]
		var dst *websocket.Conn
		if rm != nil {
			if rm.a != nil && rm.a.c == c && rm.b != nil {
				dst = rm.b.c
			} else if rm.b != nil && rm.b.c == c && rm.a != nil {
				dst = rm.a.c
			}
		}
		mu.Unlock()

		if dst != nil {
			ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := forwardRaw(ctx2, dst, msg); err != nil {
				log.Println("forward err:", err)
			}
			cancel()
		}
	}

	// 4) 片方が切れたら掃除
	mu.Lock()
	rm = rooms[roomID]
	if rm != nil {
		if rm.a != nil && rm.a.c == c {
			rm.a = nil
		}
		if rm.b != nil && rm.b.c == c {
			rm.b = nil
		}
		if rm.a == nil && rm.b == nil {
			delete(rooms, roomID)
		}
	}
	mu.Unlock()

	c.Close(websocket.StatusNormalClosure, "bye")
}

func main() {
	http.HandleFunc("/signaling", handleWS)
	log.Println("signaling ws :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
