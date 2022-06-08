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
	"time"
)

var connStr = "postgres://ethan:password@localhost/ginrummy?sslmode=disable"
var hmacSampleSecret []byte
var DB *sql.DB

func intializeDB() {
	db, err := sql.Open("postgres", connStr)
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

type jwtResponse struct {
	JWT string `json:"jwt"`
}

type errorResponse struct {
    Error string `json:"error"`
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
				c.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, jwtResponse{JWT: newJWT})
			return
		} else {
			c.JSON(http.StatusBadRequest,gin.H{"error": "Authenication failed"} )
			return
		}
	}
    fmt.Println("Authenication Failed by returning no rows");
    c.JSON(http.StatusBadRequest, gin.H{"error": "Authenication Failed, Code 100"});
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

func main() {
	intilizeJWTSecret()
	intializeDB()
	defer DB.Close()
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/login", loginEndpoint)
	router.POST("/signup", signupEndpoint)
	router.Run("localhost:8080")
}
