package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"gopkg.in/olahol/melody.v1"
)

type OutputData struct {
	MessageType string `json:"messagetype"`
	Command     string `json:"command"`
	Content     string `json:"content"`
	Game        Game   `json:"game"`
}

func connectToGame(s *melody.Session) {
	LOCK.Lock()
	gameRoomName := strings.Split(s.Request.URL.Path, "/")[2]
	gameRoomStatus := getGameRoomStatus(gameRoomName)

	for _, info := range PLAYERS {
		if s.Request.URL.Path == info.URL {
			message := OutputData{MessageType: "meta", Command: "opponent", Content: strconv.Itoa(info.ID)}
			messageStr, err := json.Marshal(message)
			if err != nil {
				fmt.Println("Error: Invalid JSON 1")
				return
			}
			s.Write([]byte(messageStr))
		}
	}

	PLAYERS[s] = &PlayerInfo{ID: IDCOUNTER, URL: s.Request.URL.Path, GameRoom: gameRoomName}
	message := OutputData{MessageType: "meta", Command: "iam", Content: strconv.Itoa(PLAYERS[s].ID)}
	messageStr, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error: Invalid JSON 2")
		return
	}
	s.Write([]byte(messageStr))
	IDCOUNTER++

	//Telling other players who just joined
	message = OutputData{MessageType: "meta", Command: "opponent", Content: strconv.Itoa(PLAYERS[s].ID)}
	messageStr, err = json.Marshal(message)
	if err != nil {
		fmt.Println("Error: Invalid JSON 3")
		return
	}
	MROUTER.BroadcastFilter(messageStr, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path && PLAYERS[s].ID != PLAYERS[q].ID
	})

	//Putting Player into a room
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
		message = OutputData{MessageType: "meta", Command: "gameroomstatus", Content: "filled", Game: *GAMES[gameRoomName]}
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

type InputFormat struct {
	MessageType string     `json:"messagetype"`
	Command     string     `json:"command"`
	Content     string     `json:"content"`
	Player      PlayerInfo `json:"playerinfo"`
	Card        Card       `json:"card"`
}

func drawCardDeck(game Game, playerNum int) (Game, error) {
	var deck []Card
	var card Card
	deck = *game.Deck
	card, deck = PopCardStack(game.Deck)
	var playerHand []Card
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

func drawCardDiscard (game Game, playerNum int) (Game, error) {
  if len(*game.DiscardPile) == 0{
    return game, errors.New("Wasnt able to take from empty discard pile")
  }
  var discard []Card;
  var card Card;
  card, discard = PopCardStack(game.DiscardPile);
	var playerHand []Card
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
	game.DiscardPile = &discard
	return game, nil
}

func filterCards(cards []Card, validCard func(Card) bool) *[]Card {
	var newCards []Card
	for _, card := range cards {
		if validCard(card) {
			newCards = append(newCards, card)
		}
	}
	return &newCards
}

func removeCard(cards []Card, cardToRemove Card) ([]Card, error){
  var newCards [] Card;
  for _, card := range cards{
    if (notEqualCards(card, cardToRemove)){
      newCards = append(newCards, card)
    }
  }
  return newCards, nil
}

func equalCards(card1 Card, card2 Card) bool {
  if card1.Rank != card2.Rank{
    return false
  }
  if card1.Suit != card2.Suit {
    return false
  }
  return true
}

func notEqualCards(card1 Card, card2 Card) bool {
  return !equalCards(card1, card2);
}

func discardCard(game Game, playerNum int, card Card) (Game, error) {
	//putting it on the discard pile
	discard := *game.DiscardPile
	discard = append(discard, card)
	game.DiscardPile = &discard

	//taking it off the players hand
	if playerNum == game.Player1.ID {
    hand, err := removeCard(*game.Player1hand, card)
    if err != nil {
      fmt.Println("Error occured during execution", err.Error())
    }
		game.Player1hand = &hand

	} else if playerNum == game.Player2.ID {
    hand, err := removeCard(*game.Player2hand, card)
    if err != nil {
      fmt.Println("Error occured during execution", err.Error())
    }
		game.Player2hand = &hand
	} else {
		return game, errors.New("The Player ID given is not a current player")
	}
	return game, nil
}


func handleGameMoves(s *melody.Session, msg []byte) {
	LOCK.Lock()
	var response OutputData
	response.MessageType = "game"
	response.Command = "gameupdate"
	var input InputFormat
	err := json.Unmarshal(msg, &input)
	if err != nil {
		fmt.Println("Error: Unable to parse json message", err.Error())
		return
	}
	//for drawing a card from the stack
	if input.Command == "draw" {
		if input.Content == "stack" {
			var game Game
			game = *GAMES[input.Player.GameRoom]
			game, err = drawCardDeck(game, input.Player.ID)
			game.Status = WaitDiscard
			GAMES[input.Player.GameRoom] = &game
			response.Game = *GAMES[input.Player.GameRoom]
    } else if input.Content == "discard" {
			var game Game
			game = *GAMES[input.Player.GameRoom]
      game, err = drawCardDiscard(game, input.Player.ID)
      game.Status = WaitDiscard
			GAMES[input.Player.GameRoom] = &game
			response.Game = *GAMES[input.Player.GameRoom]
    }
	} else if input.Command == "discard" {
		var game Game
		game = *GAMES[input.Player.GameRoom]
		game, err := discardCard(game, input.Player.ID, input.Card)
		if err != nil {
			fmt.Println("Error: There was a issue with the discardCard", err.Error())
			return
		}
		//set it to the beginning
		game.Status = BegTurn
		//give the other player their turn
		if game.Turn == game.Player1.ID {
			game.Turn = game.Player2.ID
		} else if game.Turn == game.Player2.ID {
			game.Turn = game.Player1.ID
		}
		GAMES[input.Player.GameRoom] = &game
		response.Game = *GAMES[input.Player.GameRoom]
  } else if input.Command == "gameover"{
		var game Game
		game = *GAMES[input.Player.GameRoom]
    game.Status = GameOver;
    player1HandScore, player1HandOrdered := findHandScore(*game.Player1hand)
    player2HandScore, player2HandOrdered := findHandScore(*game.Player2hand)
    fmt.Println(player1HandOrdered, player2HandOrdered)
    flatternOrderedP1Hand := flattenSets(player1HandOrdered)
    flatternOrderedP2Hand := flattenSets(player2HandOrdered)
    game.Player1hand = &flatternOrderedP1Hand
    game.Player2hand = &flatternOrderedP2Hand
		GAMES[input.Player.GameRoom] = &game
		response.Game = *GAMES[input.Player.GameRoom]
    response.Command = GameOver;
    response.Content = strconv.Itoa(player1HandScore) + " " + strconv.Itoa(player2HandScore)
  }

	result, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error: cannot parse the response")
	}
	MROUTER.BroadcastFilter(result, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path
	})
	LOCK.Unlock()
}
