package main

type Suit string

const (
	Spades   Suit = "spades"
	Clubs         = "clubs"
	Hearts        = "hearts"
	Diamonds      = "diamonds"
)

var (
	suitMap = map[string]Suit{
		"spades":   Spades,
		"clubs":    Clubs,
		"hearts":   Hearts,
		"diamonds": Diamonds,
	}
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

var (
	rankMap = map[int]Rank{
		1:  Ace,
		2:  Two,
		3:  Three,
		4:  Four,
		5:  Five,
		6:  Six,
		7:  Seven,
		8:  Eight,
		9:  Nine,
		10: Ten,
		11: Jack,
		12: Queen,
		13: King,
	}
)

type GameRoomStatus string

const (
	nonexistent GameRoomStatus = "nonexistent"
	freespot                   = "freespot"
	filled                     = "filled"
)

type Card struct {
	Rank Rank `json:"rank"`
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
	Status      GameStatus `json:"status"`
	Name        string     `json:"name"`
}

type PlayerInfo struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	GameRoom string `json:"gameroom"`
}


