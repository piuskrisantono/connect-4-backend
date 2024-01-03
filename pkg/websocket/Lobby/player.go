package websocketlobby

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string `json:"id"`
	Conn     *websocket.Conn
	Lobby    *Lobby
	Username string `json:"username"`
}

type PlayerDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type PlayersMessage struct {
	Type    string      `json:"type"`
	Players []PlayerDTO `json:"players"`
}

type BattleConfirmationMessage struct {
	Type      string    `json:"type"`
	BattleId  string    `json:"battleId"`
	PlayerOne PlayerDTO `json:"playerOne"`
}

type BattleMessage struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerId"`
}

type BattleRegistration struct {
	PlayerOneId string
	PlayerTwoId string
}

func (c *Player) Read() {
	defer func() {
		c.Lobby.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, bytes, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("error on read message", err)
			return
		}

		battleMessage := BattleMessage{}

		errorParsing := json.Unmarshal(bytes, &battleMessage)

		if errorParsing != nil {
			log.Println("error on parsing", errorParsing)
			return
		}

		switch battleMessage.Type {
		case "battle":
			c.Lobby.BattleRegistration <- BattleRegistration{c.ID, battleMessage.PlayerId}
			break
		default:
		}
	}
}
