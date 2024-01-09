package websocketlobby

import (
	"encoding/json"
	"fmt"
	"log"
)

type Lobby struct {
	Register      chan *Player
	Unregister    chan *Player
	Players       []*Player
	Battles       map[string]*BattleRoom
	BattleMessage chan *BattleMessage
}

type BattleRoom struct {
	PlayerOne *Player `json:"playerOne"`
	PlayerTwo *Player `json:"playerTwo"`
}

type BattleFillRequest struct {
	Type     string `json:"type"`
	BattleId string `json:"battleId"`
	ColIndex int    `json:"colIndex"`
}

type BattleInfoResponse struct {
	Type       string     `json:"type"`
	BattleId   string     `json:"battleId"`
	BattleRoom BattleRoom `json:"battleRoom"`
}

type PlayersResponse struct {
	Type    string    `json:"type"`
	Players []*Player `json:"players"`
}

type BattleConfirmationResponse struct {
	Type      string `json:"type"`
	BattleId  string `json:"battleId"`
	PlayerOne Player `json:"playerOne"`
}

type BattleFillResponse struct {
	Type     string `json:"type"`
	ColIndex int    `json:"colIndex"`
}

func NewLobby() *Lobby {
	return &Lobby{
		Register:      make(chan *Player),
		Unregister:    make(chan *Player),
		Players:       []*Player{},
		Battles:       make(map[string]*BattleRoom),
		BattleMessage: make(chan *BattleMessage),
	}
}

func (lobby *Lobby) Start() {
	for {
		select {
		case newPlayer := <-lobby.Register:
			lobby.Players = append(lobby.Players, newPlayer)
			lobby.broadcastPlayers()
		case dcPlayer := <-lobby.Unregister:
			newPlayers := []*Player{}
			for _, player := range lobby.Players {
				if player.ID != dcPlayer.ID {
					newPlayers = append(newPlayers, player)
				}
			}
			lobby.Players = newPlayers
			lobby.broadcastPlayers()
		case message := <-lobby.BattleMessage:
			currentPlayerId := message.PlayerId

			switch message.Type {
			case "battle":
				var playerTwoId string
				errorParsing := json.Unmarshal(message.Content, &playerTwoId)
				if errorParsing != nil {
					log.Println("error on parsing battle message", errorParsing)
					continue
				}
				battleRoom := BattleRoom{}
				for _, player := range lobby.Players {
					if player.ID == currentPlayerId {
						battleRoom.PlayerOne = player
					} else if player.ID == playerTwoId {
						battleRoom.PlayerTwo = player
					}

					if battleRoom.PlayerOne != nil && battleRoom.PlayerTwo != nil {
						break
					}
				}

				battleId := currentPlayerId + playerTwoId

				lobby.Battles[battleId] = &battleRoom

				battleConfirmationMessage := BattleConfirmationResponse{
					Type:     "confirmation",
					BattleId: battleId,
					PlayerOne: Player{
						ID:       battleRoom.PlayerOne.ID,
						Username: battleRoom.PlayerOne.Username,
					},
				}

				if err := battleRoom.PlayerTwo.Conn.WriteJSON(battleConfirmationMessage); err != nil {
					fmt.Println(err)
				}
			case "accept":
				var battleId string
				errorParsing := json.Unmarshal(message.Content, &battleId)
				if errorParsing != nil {
					log.Println("error on parsing accept message", errorParsing)
					continue
				}
				battleRoom := lobby.Battles[battleId]
				battleRoom.PlayerOne.Conn.WriteJSON(BattleInfoResponse{"accept", battleId, *battleRoom})
			case "decline", "over":
				var battleId string
				errorParsing := json.Unmarshal(message.Content, &battleId)
				if errorParsing != nil {
					log.Println("error on parsing decline message", errorParsing)
					continue
				}
				delete(lobby.Battles, battleId)
			case "fill":
				battleFillRequest := BattleFillRequest{}
				errorParsing := json.Unmarshal(message.Content, &battleFillRequest)
				if errorParsing != nil {
					log.Println("error on parsing fill message", errorParsing)
					continue
				}
				battleRoom := lobby.Battles[battleFillRequest.BattleId]
				connectionToSend := battleRoom.PlayerOne.Conn
				if currentPlayerId == battleRoom.PlayerOne.ID {
					connectionToSend = battleRoom.PlayerTwo.Conn
				}
				connectionToSend.WriteJSON(BattleFillResponse{"fill", battleFillRequest.ColIndex})
			default:
			}
		}
	}
}

func (lobby *Lobby) broadcastPlayers() {
	for _, player := range lobby.Players {
		playersMessage := PlayersResponse{
			Type:    "players",
			Players: lobby.Players,
		}
		player.Conn.WriteJSON(playersMessage)
	}
}
