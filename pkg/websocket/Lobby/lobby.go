package websocketlobby

import "fmt"

type Lobby struct {
	Register           chan *Player
	Unregister         chan *Player
	Players            []*Player
	BattleRegistration chan BattleRegistration
	Battles            map[string]*BattleRoom
	BattleAccept       chan string
	BattleDecline      chan string
	BattleFill         chan BattleFill
}

type BattleRoom struct {
	PlayerOne *Player `json:"playerOne"`
	PlayerTwo *Player `json:"playerTwo"`
}

type BattleInfo struct {
	Type       string     `json:"type"`
	BattleId   string     `json:"battleId"`
	BattleRoom BattleRoom `json:"battleRoom"`
}

func NewLobby() *Lobby {
	return &Lobby{
		Register:           make(chan *Player),
		Unregister:         make(chan *Player),
		Players:            []*Player{},
		BattleRegistration: make(chan BattleRegistration),
		Battles:            make(map[string]*BattleRoom),
		BattleAccept:       make(chan string),
		BattleDecline:      make(chan string),
		BattleFill:         make(chan BattleFill),
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
		case battleRegistration := <-lobby.BattleRegistration:
			battleRoom := BattleRoom{}
			for _, player := range lobby.Players {
				if player.ID == battleRegistration.PlayerOneId {
					battleRoom.PlayerOne = player
				} else if player.ID == battleRegistration.PlayerTwoId {
					battleRoom.PlayerTwo = player
				}

				if battleRoom.PlayerOne != nil && battleRoom.PlayerTwo != nil {
					break
				}
			}

			battleId := battleRegistration.PlayerOneId + battleRegistration.PlayerTwoId

			lobby.Battles[battleId] = &battleRoom

			battleConfirmationMessage := BattleConfirmationMessage{
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
		case battleId := <-lobby.BattleAccept:
			battleRoom := lobby.Battles[battleId]
			battleRoom.PlayerOne.Conn.WriteJSON(BattleInfo{"accept", battleId, *battleRoom})
		case battleId := <-lobby.BattleDecline:
			delete(lobby.Battles, battleId)
		case battleFill := <-lobby.BattleFill:
			battleRoom := lobby.Battles[battleFill.BattleId]
			connectionToSend := battleRoom.PlayerOne.Conn
			if battleFill.PlayerId == battleRoom.PlayerOne.ID {
				connectionToSend = battleRoom.PlayerTwo.Conn
			}
			connectionToSend.WriteJSON(BattleFillMessage{"fill", "", battleFill.ColIndex})
		}
	}
}

func (lobby *Lobby) broadcastPlayers() {
	for _, player := range lobby.Players {
		playersMessage := PlayersMessage{
			Type:    "players",
			Players: lobby.Players,
		}
		player.Conn.WriteJSON(playersMessage)
	}
}
