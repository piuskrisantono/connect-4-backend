package websocketlobby

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string          `json:"id"`
	Conn     *websocket.Conn `json:"-"`
	Lobby    *Lobby          `json:"-"`
	Username string          `json:"username"`
}

// type Message struct {
// 	Type string `json:"type"`
// 	Content json.RawMessage `json:"content"`
// }

type PlayersMessage struct {
	Type    string    `json:"type"`
	Players []*Player `json:"players"`
}

type BattleConfirmationMessage struct {
	Type      string `json:"type"`
	BattleId  string `json:"battleId"`
	PlayerOne Player `json:"playerOne"`
}

type BattleMessage struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerId"`
}

type BattleAnswerMessage struct {
	Type     string `json:"type"`
	BattleId string `json:"battleId"`
}

type BattleRegistration struct {
	PlayerOneId string
	PlayerTwoId string
}

type BattleFillMessage struct {
	Type     string `json:"type"`
	BattleId string `json:"battleId"`
	ColIndex int    `json:"colIndex"`
}

type BattleFill struct {
	PlayerId string `json:"playerId"`
	BattleId string `json:"battleId"`
	ColIndex int    `json:"colIndex"`
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

		message := string(bytes)

		if strings.Contains(message, "\"type\":\"battle\"") {
			battleMessage := BattleMessage{}

			errorParsing := json.Unmarshal(bytes, &battleMessage)

			if errorParsing != nil {
				log.Println("error on parsing", errorParsing)
				continue
			}
			c.Lobby.BattleRegistration <- BattleRegistration{c.ID, battleMessage.PlayerId}
		} else if strings.Contains(message, "\"type\":\"accept\"") {
			battleAnswerMessage := BattleAnswerMessage{}
			errorParsing := json.Unmarshal(bytes, &battleAnswerMessage)
			if errorParsing != nil {
				log.Println("error on parsing", errorParsing)
				continue
			}
			c.Lobby.BattleAccept <- battleAnswerMessage.BattleId
		} else if strings.Contains(message, "\"type\":\"decline\"") {
			battleAnswerMessage := BattleAnswerMessage{}
			errorParsing := json.Unmarshal(bytes, &battleAnswerMessage)
			if errorParsing != nil {
				log.Println("error on parsing", errorParsing)
				continue
			}
			c.Lobby.BattleDecline <- battleAnswerMessage.BattleId
		} else if strings.Contains(message, "\"type\":\"fill\"") {
			battleFillMessage := BattleFillMessage{}
			errorParsing := json.Unmarshal(bytes, &battleFillMessage)
			if errorParsing != nil {
				log.Println("error on parsing", errorParsing)
				continue
			}
			c.Lobby.BattleFill <- BattleFill{c.ID, battleFillMessage.BattleId, battleFillMessage.ColIndex}
		}
	}
}
