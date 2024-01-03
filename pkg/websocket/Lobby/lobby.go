package websocketlobby

import "fmt"

type Lobby struct {
	Register           chan *Player
	Unregister         chan *Player
	Players            []*Player
	BattleRegistration chan BattleRegistration
	Battles            map[string]*BattleRoom
}

type BattleRoom struct {
	PlayerOne *Player
	PlayerTwo *Player
}

func NewLobby() *Lobby {
	return &Lobby{
		Register:           make(chan *Player),
		Unregister:         make(chan *Player),
		Players:            []*Player{},
		BattleRegistration: make(chan BattleRegistration),
		Battles:            make(map[string]*BattleRoom),
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
				Type:      "confirmation",
				BattleId:  battleId,
				PlayerOne: PlayerDTO{battleRoom.PlayerOne.ID, battleRoom.PlayerOne.Username},
			}

			if err := battleRoom.PlayerTwo.Conn.WriteJSON(battleConfirmationMessage); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (lobby *Lobby) broadcastPlayers() {
	for _, player := range lobby.Players {
		playersMessage := PlayersMessage{
			Type:    "players",
			Players: lobby.generatePlayerDTOS(),
		}
		player.Conn.WriteJSON(playersMessage)
	}
}

func (lobby *Lobby) generatePlayerDTOS() []PlayerDTO {
	playerDTOS := []PlayerDTO{}
	for _, player := range lobby.Players {
		playerDTOS = append(playerDTOS, PlayerDTO{player.ID, player.Username})
	}
	return playerDTOS
}
