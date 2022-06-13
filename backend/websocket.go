package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"gopkg.in/olahol/melody.v1"
)

func connectToGame(s *melody.Session) {
	LOCK.Lock()
	for _, info := range PLAYERS {
		if s.Request.URL.Path == info.URL {
			message := WSMetaJSONFormat{MessageType: "meta", Command: "opponent", Content: strconv.Itoa(info.ID)}
			messageStr, err := json.Marshal(message)
			if err != nil {
				fmt.Println("Error: Invalid JSON 1")
				return
			}
			s.Write([]byte(messageStr))
		}
	}

	gameRoomName := strings.Split(s.Request.URL.Path, "/")[2]
	PLAYERS[s] = &PlayerInfo{ID: IDCOUNTER, URL: s.Request.URL.Path, GameRoom: gameRoomName}
	message := WSMetaJSONFormat{MessageType: "meta", Command: "iam", Content: strconv.Itoa(PLAYERS[s].ID)}
	messageStr, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error: Invalid JSON 2")
		return
	}
	s.Write([]byte(messageStr))
	IDCOUNTER++

	//Telling other players who just joined
	message = WSMetaJSONFormat{MessageType: "meta", Command: "opponent", Content: strconv.Itoa(PLAYERS[s].ID)}
	messageStr, err = json.Marshal(message)
	if err != nil {
		fmt.Println("Error: Invalid JSON 3")
		return
	}
	MROUTER.BroadcastFilter(messageStr, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path && PLAYERS[s].ID != PLAYERS[q].ID
	})

	//Putting Player into a room
	gameRoomStatus := getGameRoomStatus(gameRoomName)
	if gameRoomStatus == "nonexistent" {
    Deck := makeDeck()
		Player1hand, Deck := MakeHand(Deck)
		Player2hand, Deck := MakeHand(Deck)
    var card Card
    card, *Deck = PopCardStack(Deck)
    DiscardPile := []Card{card}
    GAMES[gameRoomName] = &Game{Name: PLAYERS[s].GameRoom, Turn: PLAYERS[s].ID, Player1: *PLAYERS[s], Deck: Deck, Player1hand: &Player1hand, Player2hand: &Player2hand, DiscardPile: &DiscardPile}
	} else {
		if GAMES[gameRoomName].Player1.ID == 0 {
			GAMES[gameRoomName].Player1 = *PLAYERS[s]
		} else if GAMES[gameRoomName].Player2.ID == 0 {
			GAMES[gameRoomName].Player2 = *PLAYERS[s]
		} else {
			fmt.Println("Error: the room was full when we tried to join")
		}

	}



	//now checking if the room is full or not
	gameRoomStatus = getGameRoomStatus(gameRoomName)
	if gameRoomStatus == filled {
    GAMES[gameRoomName].Status = Starting
		message = WSMetaJSONFormat{MessageType: "meta", Command: "gameroomstatus", Content: "filled", Game: *GAMES[gameRoomName]}
		messageStr, err = json.Marshal(message)
		if err != nil {
			fmt.Println("Error: Invalid JSON 4")
			return
		}
		MROUTER.BroadcastFilter(messageStr, func(q *melody.Session) bool {
			return q.Request.URL.Path == s.Request.URL.Path
		})
	}

	LOCK.Unlock()
}

func leaveGame(s *melody.Session) {
	LOCK.Lock()

	//removing player if they belong to a game
	successfulRemoval := false
	for _, game := range GAMES {
		if game.Player1.ID == PLAYERS[s].ID {
			var newPlayer = new(PlayerInfo)
			game.Player1 = *newPlayer
			successfulRemoval = true
      if game.Player2.ID == 0 {
        delete(GAMES, game.Name)
      }
			break
		}
		if game.Player2.ID == PLAYERS[s].ID {
			var newPlayer = new(PlayerInfo)
			game.Player2 = *newPlayer
			successfulRemoval = true
      if game.Player1.ID == 0 {
        delete(GAMES, game.Name)
      }
			break
		}
	}
	if !successfulRemoval {
		fmt.Println("Error: failure to remove player: " + strconv.Itoa(PLAYERS[s].ID))
	}
	msg := []byte("disconnect " + strconv.Itoa(PLAYERS[s].ID))
	MROUTER.BroadcastFilter(msg, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path
	})
	delete(PLAYERS, s)
	LOCK.Unlock()
}

type WSGameJSONFormat struct {
  MessageType string `json:"messagetype"`
  Command string `json:"command"`
  Content string `json:"content"`
  Player PlayerInfo `json:"playerinfo"`
}

func drawCardDeck (game Game, playerNum int) (Game, error) {
  var deck []Card;
  var card Card;
  deck = *game.Deck
  card, deck = PopCardStack(game.Deck)
  var playerHand []Card;
  if playerNum == game.Player1.ID {
    playerHand = *game.Player1hand
    playerHand = append(playerHand, card)
    game.Player1hand = &playerHand
  }
  if playerNum == game.Player2.ID {
    playerHand = *game.Player2hand
    playerHand = append(playerHand, card)
    game.Player2hand = &playerHand
  }
  game.Deck = &deck
  return game, nil
}

func discardCard(game Game, playerNum int, card string) (Game, error) {
  //not finished obviously 
  return game, nil
}


func handleGameMoves(s *melody.Session, msg []byte) {
	LOCK.Lock()
  var response WSMetaJSONFormat
  var input WSGameJSONFormat
  err := json.Unmarshal(msg, &input)
  if err != nil{
    fmt.Println("Error: Unable to parse json message", err.Error())
    return
  }
  fmt.Println(input)
  //for drawing a card from the stack 
  if input.Command == "draw" {
    if input.Content == "stack"{
      var game Game
      game = *GAMES[input.Player.GameRoom]
      game, err = drawCardDeck(game, input.Player.ID)
      game.Status = WaitDiscard
      GAMES[input.Player.GameRoom] = &game
      response.Game = *GAMES[input.Player.GameRoom]
      response.MessageType = "game"
      response.Command = "gameupdate"
    }
  }

  if input.Command == "discard" {

  }

  result, err:= json.Marshal(response)
  if err != nil {
    fmt.Println("Error: cannot parse the response")
  }
	MROUTER.BroadcastFilter(result, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path
	})
	LOCK.Unlock()
}
