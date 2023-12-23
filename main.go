package main

import (
	"fmt"
	"net/http"

	"github.com/piuskrisantono/connect-4-backend.git/pkg/websocket"
)

func main() {
	fmt.Println("chat app v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}

func setupRoutes() {
	pool := websocket.NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}
