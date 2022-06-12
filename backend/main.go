package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/olahol/melody.v1"
)

var CONNSTR = "postgres://ethan:password@localhost/ginrummy?sslmode=disable"
var hmacSampleSecret []byte
var DB *sql.DB
var MROUTER *melody.Melody
var LOCK *sync.Mutex
var PLAYERS map[*melody.Session]*PlayerInfo
var IDCOUNTER int
var GAMES map[string]*Game

var Ranks = []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}
var Suits = []Suit{Spades, Clubs, Hearts, Diamonds}

func MakeHand (Deck *[]Card) (OutHand []Card, OutDeck *[]Card){
  DeckInstance := *Deck;
  var hand []Card;
  var card Card;
  for i:=0; i<10; i++ {
    card, DeckInstance = DeckInstance[len(DeckInstance)-1], DeckInstance[:len(DeckInstance)-1]
    hand = append(hand, card)
  }
  return hand, &DeckInstance
}




func PushCardStack(a []Card, x Card) ([]Card){
  return append(a, x)
}

func PopCardStack(a *[]Card) (Card, []Card){
  val := *a
  return val[len(val)-1], val[:len(val)-1]
}

func makeDeck() *[]Card {
  Deck := *new([]Card)
	for _, suit := range Suits {
		for _, rank := range Ranks {
			thisCard := Card{Rank: rank, Suit: suit}
			Deck = append(Deck, thisCard)
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(Deck), func(i, j int) { Deck[i], Deck[j] = Deck[j], Deck[i] })
	return &Deck
}

func intializeDB() {
	db, err := sql.Open("postgres", CONNSTR)
	if err != nil {
		fmt.Println("ERROR!! Database failed to intilize")
		return
	}
	DB = db
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type signupInput struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func signupEndpoint(c *gin.Context) {
	var input signupInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "Signup Failed: Incorrect Input")
		return
	}
	hashedpassword, err := HashPassword(input.Password)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "Signup Failed: Password unhashable")
		return
	}
	_, err = DB.Exec("insert into users (name, username, password, email) values ($1, $2, $3, $4);", input.Name, input.Username, hashedpassword, input.Email)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "Signup Failed: Database Rejection")
		return
	}
	c.JSON(http.StatusOK, "Account Created")
}

type loginEndpointInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func loginEndpoint(c *gin.Context) {
	var input loginEndpointInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, err := DB.Query("select email, password from users where username = $1", input.Username)
	if err != nil {
		fmt.Println(gin.H{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()
	for rows.Next() {
		var hashedpassword string
		var email string
		err := rows.Scan(&email, &hashedpassword)
		if err != nil {
			fmt.Println(gin.H{"error": err.Error()})
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if CheckPasswordHash(input.Password, hashedpassword) {
			newJWT, err := getJWT(input.Username, email)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"jwt": newJWT})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authenication failed"})
			return
		}
	}
	fmt.Println("Authenication Failed by returning no rows")
	c.JSON(http.StatusBadRequest, gin.H{"error": "Authenication Failed, Code 100"})
	return
}

func intilizeJWTSecret() {
	hmacSampleSecret = []byte("Avacado")
}

func getJWT(username string, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"email":    email,
		"exp":      time.Now().Unix() + 7200,
	})
	return token.SignedString(hmacSampleSecret)
}

func validateAndDecryptJWT(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok1 := claims["username"].(string)
		email, ok2 := claims["email"].(string)

		if ok1 && ok2 {
			return username, email, nil
		} else {
			return "", "", errors.New("Types of email and password were wrong")
		}
	} else {
		return "", "", errors.New(err.Error())
	}

}
func handleWSchannel(c *gin.Context) {
	MROUTER.HandleRequest(c.Writer, c.Request)
}

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

func handleGameMoves(s *melody.Session, msg []byte) {
	LOCK.Lock()
	MROUTER.BroadcastFilter(msg, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path
	})
	LOCK.Unlock()
}

func intializeWSvars() {
	MROUTER = melody.New()
	PLAYERS = make(map[*melody.Session]*PlayerInfo)
	LOCK = new(sync.Mutex)
	IDCOUNTER = 1
	GAMES = make(map[string]*Game)
}

func gameRoomQuery(c *gin.Context) {
	var input gameQueryInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "Failed: incorrect input")
		return
	}
	c.JSON(http.StatusOK, gin.H{"gameroomstatus": getGameRoomStatus(input.GameName)})
}

func getGameRoomStatus(val string) GameRoomStatus {
	thisGame, present := GAMES[val]
	if !present {
		return nonexistent
	}
	if thisGame.Player1.ID == 0 || thisGame.Player2.ID == 0 {
		return freespot
	}
	return filled
}

func main() {
	intilizeJWTSecret()
	intializeDB()
	defer DB.Close()
	intializeWSvars()
	router := gin.Default()
	router.GET("/channel/:name/play", handleWSchannel)
	MROUTER.HandleConnect(connectToGame)
	MROUTER.HandleDisconnect(leaveGame)
	MROUTER.HandleMessage(handleGameMoves)
	router.Use(cors.Default())
	router.POST("/login", loginEndpoint)
	router.POST("/signup", signupEndpoint)
	router.GET("/gameRoomQuery", gameRoomQuery)
	router.Run("localhost:8080")
}
