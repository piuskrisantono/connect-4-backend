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

type Message struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type BattleFillRequest struct {
	Type     string `json:"type"`
	BattleId string `json:"battleId"`
	ColIndex int    `json:"colIndex"`
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

		message := Message{}

		errorParsingMessage := json.Unmarshal(bytes, &message)
		if errorParsingMessage != nil {
			log.Println("error on parsing", errorParsingMessage)
			continue
		}

		switch message.Type {
		case "battle":
			var playerId string
			errorParsing := json.Unmarshal(message.Content, &playerId)
			if errorParsing != nil {
				log.Println("error on parsing battle message", errorParsing)
				continue
			}
			p.Lobby.BattleRegistration <- BattleRegistration{p.ID, playerId}
		case "accept":
			var battleId string
			errorParsing := json.Unmarshal(message.Content, &battleId)
			if errorParsing != nil {
				log.Println("error on parsing accept message", errorParsing)
				continue
			}
			p.Lobby.BattleAccept <- battleId
		case "decline":
			var battleId string
			errorParsing := json.Unmarshal(message.Content, &battleId)
			if errorParsing != nil {
				log.Println("error on parsing decline message", errorParsing)
				continue
			}
			p.Lobby.BattleDecline <- battleId
		case "fill":
			battleFillRequest := BattleFillRequest{}
			errorParsing := json.Unmarshal(message.Content, &battleFillRequest)
			if errorParsing != nil {
				log.Println("error on parsing fill message", errorParsing)
				continue
			}
			p.Lobby.BattleFill <- BattleFill{p.ID, battleFillRequest.BattleId, battleFillRequest.ColIndex}
		default:
		}
	}
}
