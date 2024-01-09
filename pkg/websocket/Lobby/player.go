package websocketlobby

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string          `json:"id"`
	Conn     *websocket.Conn `json:"-"`
	Lobby    *Lobby          `json:"-"`
	Username string          `json:"username"`
}

type BattleMessage struct {
	Type     string          `json:"type"`
	Content  json.RawMessage `json:"content"`
	PlayerId string
}

func (p *Player) Read() {
	defer func() {
		p.Lobby.Unregister <- p
		p.Conn.Close()
	}()

	for {
		_, bytes, err := p.Conn.ReadMessage()
		if err != nil {
			log.Println("error on read message", err)
			return
		}

		battleMessage := BattleMessage{}

		errorParsingMessage := json.Unmarshal(bytes, &battleMessage)
		if errorParsingMessage != nil {
			log.Println("error on parsing", errorParsingMessage)
			continue
		}

		battleMessage.PlayerId = p.ID

		p.Lobby.BattleMessage <- &battleMessage
	}
}
