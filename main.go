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

	// "github.com/pion/webrtc/v3"
	// "github.com/urabexon/WasmP2PChat/"
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

func shortHash(now time.Time) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(now.String())); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:7], nil
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

	// fmt.Println("Hello, World!")
	select {}
}
