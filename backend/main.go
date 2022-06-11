package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"sync"
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
  freespot = "freespot"
  filled = "filled"
)


var Ranks = []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}
var Suits = []Suit{Spades, Clubs, Hearts, Diamonds}

type Card struct {
	rank Rank
	suit Suit
}

func makeDeck() *[52]Card {
	var Deck [52]Card
	for i, suit := range Suits {
		for j, rank := range Ranks {
			thisCard := new(Card)
			thisCard.rank = rank
			thisCard.suit = suit
			Deck[i*j] = *thisCard
		}
	}
	return &Deck
}

type Game struct {
	Player1     PlayerInfo
	Player2     PlayerInfo
	Deck        *[52]Card
	Player1hand *[]Card
	Player2hand *[]Card
	DiscardPile *[]Card
}

func intializeDB() {
	db, err := sql.Open("postgres", CONNSTR)
	if err != nil {
		fmt.Println("ERROR!! Database failed to intilize")
		return
	}
	DB = db
}

//OLD STRUCT
type user struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
}

//OLD FUNCTION
func getDataExample(c *gin.Context) {
	users := []user{}
	rows, err := DB.Query("select userid, uname from users")
	if err != nil {
		fmt.Println("query failed")
		fmt.Println(err)
		c.IndentedJSON(http.StatusBadRequest, nil)
		return
	}
	defer rows.Close()
	for rows.Next() {
		user := user{Username: "", Fullname: ""}
		err := rows.Scan(&user.Username, &user.Fullname)
		if err != nil {
			fmt.Println("failed to scan row")
			c.IndentedJSON(http.StatusBadRequest, nil)
			return
		}
		users = append(users, user)

	}
	err = rows.Err()
	if err != nil {
		fmt.Println("rows somehow had an error?")
		fmt.Println(err)
		c.IndentedJSON(http.StatusBadRequest, nil)
	}
	c.IndentedJSON(http.StatusOK, users)
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

type PlayerInfo struct {
	ID       int
	URL      string
	GameRoom string
}

func connectToGame(s *melody.Session) {
	LOCK.Lock()
	for _, info := range PLAYERS {
		if s.Request.URL.Path == info.URL {
			s.Write([]byte("otherplayer " + strconv.Itoa(info.ID)))
		}
	}

	gameRoomName := strings.Split(s.Request.URL.Path, "/")[2]
	PLAYERS[s] = &PlayerInfo{ID: IDCOUNTER, URL: s.Request.URL.Path, GameRoom: gameRoomName}
	s.Write([]byte("iam " + strconv.Itoa(PLAYERS[s].ID) + " " + PLAYERS[s].URL))
	IDCOUNTER++

	//Telling other players who just joined
	msg := []byte("otherplayer " + strconv.Itoa(PLAYERS[s].ID))
	MROUTER.BroadcastFilter(msg, func(q *melody.Session) bool {
		return q.Request.URL.Path == s.Request.URL.Path
	})

	//Putting Player into a room
	gameRoomStatus := getGameRoomStatus(gameRoomName)
	if gameRoomStatus == "nonexistent" {
    Player1hand := new([]Card)
    Player2hand := new([]Card)
    DiscardPile := new([]Card)
    GAMES[gameRoomName] = &Game{Player1: *PLAYERS[s], Deck: makeDeck(), Player1hand: Player1hand, Player2hand: Player2hand, DiscardPile: DiscardPile }
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
		  msg = []byte("Game Room Filled")
      MROUTER.BroadcastFilter(msg, func(q *melody.Session) bool {
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
      break
    }
    if game.Player2.ID == PLAYERS[s].ID {
      var newPlayer = new(PlayerInfo)
      game.Player2 = *newPlayer 
      successfulRemoval = true
      break
    }
  }
  if !successfulRemoval {
    fmt.Println("Error: failure to remove player: " + strconv.Itoa(PLAYERS[s].ID));
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

type gameQueryInput struct {
	GameName string `json:"gameroom" binding:"required"`
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
		return  freespot
 	}
	return filled 
}


/*
Game Logic Ideas: 

- We need a instruction set for the game 

The Game needs to start when two players are in a room and stop when someone has disconnected. I need a formal langauge for this kinda 
Possible moves:
- every turn starts with a choice to knock or not (dpending on if you can) 
- every turn then has to decide to pick a card from discard or from deck 
- then the player moves around their cards and discards one ending their turn , they could also place that card upside down which could cause issues but heyik 

First I have to sense when room is full pretty easy 

Then I have to come up with different types of messages 

*/

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
