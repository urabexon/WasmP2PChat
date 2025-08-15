package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
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

var queue = make(chan *websocket.Conn, 128)

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func handleWS(w http.ResponseWriter, r *http.Request) {
    // c, err := websocket.Accept(w, r, nil)
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
	    OriginPatterns: []string{"*"},
	})
    if err != nil {
        // log.Println("ws accept:", err)
		log.Println("ws accept:", err, "origin:", r.Header.Get("Origin"))
        return
    }
    defer c.Close(websocket.StatusInternalError, "server error")

    var req mmReqMsg
    if err := wsjson.Read(r.Context(), c, &req); err != nil {
        log.Println("read:", err)
        return
    }

    // 先に待っている人がいるか確認する
    select {
    case partner := <-queue:
        // ペア成立
        roomID := shortID()
        now := time.Now()
        res1 := mmResMsg{Type: "MATCH", RoomID: roomID, UserID: req.UserID, CreatedAt: now}
        res2 := mmResMsg{Type: "MATCH", RoomID: roomID, UserID: "peer", CreatedAt: now}

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := wsjson.Write(ctx, c, res1); err != nil {
            log.Println("write1:", err)
            return
        }
        if err := wsjson.Write(ctx, partner, res2); err != nil {
            log.Println("write2:", err)
            return
        }
        c.Close(websocket.StatusNormalClosure, "ok")
        partner.Close(websocket.StatusNormalClosure, "ok")

    default:
        // 誰もいなければ待機に入る
        queue <- c
        <-r.Context().Done()
    }
}

func main() {
	http.HandleFunc("/matchmaking", handleWS)
	log.Println("matchmaking ws :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
