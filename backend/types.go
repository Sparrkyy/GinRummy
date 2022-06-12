package main

type Suit int

const (
	Spades Suit = iota
	Clubs
	Hearts
	Diamonds
)

type Rank int

const (
	Ace   Rank = 1
	Two        = 2
	Three      = 3
	Four       = 4
	Five       = 5
	Six        = 6
	Seven      = 7
	Eight      = 8
	Nine       = 9
	Ten        = 10
	Jack       = 11
	Queen      = 12
	King       = 13
)

type GameRoomStatus string

const (
	nonexistent GameRoomStatus = "nonexistent"
	freespot                   = "freespot"
	filled                     = "filled"
)

type Card struct {
	Rank Rank `json:"card"`
	Suit Suit `json:"suit"`
}

type GameStatus string

const (
	Starting    = "starting"
	BegTurn     = "begturn"
	WaitDiscard = "waitdiscard"
)

type Game struct {
	Turn        int        `json:"turn"`
	Player1     PlayerInfo `json:"player1"`
	Player2     PlayerInfo `json:"player2"`
	Deck        *[]Card    `json:"deck"`
	Player1hand *[]Card    `json:"player1hand"`
	Player2hand *[]Card    `json:"player2hand"`
	DiscardPile *[]Card    `json:"discardpile"`
	Status      GameStatus `json:"gamestatus"`
  Name string `json:"name"`
}

type PlayerInfo struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	GameRoom string `json:"gameroom"`
}

type WSMetaJSONFormat struct {
	MessageType string `json:"messagetype"`
	Command     string `json:"command"`
	Content     string `json:"content"`
	Game        Game   `json:"game"`
}

type gameQueryInput struct {
	GameName string `json:"gameroom" binding:"required"`
}
