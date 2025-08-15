package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var (
	mu      sync.Mutex
	waiting *websocket.Conn // いま待っている人（0か1）
)

func handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
	       OriginPatterns: []string{"*"},
	   })
	if err != nil {
		log.Println("ws accept:", err, "origin:", r.Header.Get("Origin"))
		return
	}

	var req mmReqMsg
	if err := wsjson.Read(r.Context(), c, &req); err != nil {
		log.Println("read:", err)
		return
	}
	log.Printf("joined user=%s\n", req.UserID)

	// 待機者がいなければ自分を待機にセットして終了
	mu.Lock()
	if waiting == nil {
		waiting = c
		mu.Unlock()
		// 接続が切れるまで待つ（ページ閉じたら自動でDoneになる）
		<-r.Context().Done()
		mu.Lock()
		if waiting == c {
			waiting = nil
		}
		mu.Unlock()
		c.Close(websocket.StatusNormalClosure, "left")
        return
	}

	// 待機者がいればペアリング
	partner := waiting
	waiting = nil
	mu.Unlock()

	roomID := shortID()
	now := time.Now()
	res1 := mmResMsg{Type: "MATCH", RoomID: roomID, UserID: req.UserID, CreatedAt: now}
	res2 := mmResMsg{Type: "MATCH", RoomID: roomID, UserID: "peer", CreatedAt: now}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 今来た人
	if err := wsjson.Write(ctx, c, res1); err != nil {
		log.Println("write to current:", err)
		return
	}
	// 待機していた人
	if err := wsjson.Write(ctx, partner, res2); err != nil {
		log.Println("write to partner:", err)
		return
	}

	log.Printf("matched room=%s\n", roomID)
	// ここでサーバ側からは閉じない（クライアントが適宜閉じる）
}

func main() {
	http.HandleFunc("/matchmaking", handleWS)
	log.Println("matchmaking ws :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
