package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
	//"errors"
)

var connStr = "postgres://ethan:password@localhost/ginrummy?sslmode=disable"
var DB * sql.DB

func intializeDB(){
	db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Println("ERROR!! Database failed to intilize");
        return
    }
    DB = db;
}

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

type user struct {
    Username string `json:"username"`
    Fullname string `json:"fullname"`
}


func getDataExample(c *gin.Context) {
    users := []user{};
	rows, err := DB.Query("select userid, uname from users")
	if err != nil {
		fmt.Println("query failed")
		fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, nil);
        return 
	}
	defer rows.Close()
	for rows.Next() {
        user := user{Username: "", Fullname: ""}
		err := rows.Scan(&user.Username, &user.Fullname)
		if err != nil {
			fmt.Println("failed to scan row")
            c.IndentedJSON(http.StatusBadRequest, nil);
            return 
		}
        users = append(users, user);

	}
	err = rows.Err()
	if err != nil {
		fmt.Println("rows somehow had an error?")
		fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, nil);
	}
    c.IndentedJSON(http.StatusOK, users);
}

type loginEndpointInput struct {
    Username string `json:"username"`
    Password string `json:"password"`
}


func loginEndpoint(c *gin.Context) {
    var input loginEndpointInput;
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    fmt.Println(input);
    //stmt, err := DB.Prepare("select password from users where username = '?'");
    rows, err := DB.Query("select password from users where username = $1", input.Username);
    if err != nil {
        fmt.Println(gin.H{"error": err.Error()})
        c.JSON(http.StatusBadRequest, "Authenication Failed")
        return
    }

	defer rows.Close()

    for rows.Next() {
        var password string
        err := rows.Scan(&password)
        if err != nil {
            fmt.Println(gin.H{"error": err.Error()})
            c.JSON(http.StatusBadRequest, "Authenication Failed")
            return
        }
        if input.Password == password {
            c.JSON(http.StatusOK, "Logged in")
            return
        } else {
            c.JSON(http.StatusBadRequest, "Authenication Failed")
            return
        }
    }
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
    Name string `json:"name"` 
    Username string `json:"username"` 
    Password string `json:"password"` 
    Email string `json:"email"` 
}

func signupEndpoint(c *gin.Context) {
    var input signupInput;
    err := c.ShouldBindJSON(&input); 
    if err != nil { fmt.Println(err.Error()); c.JSON(http.StatusBadRequest, "Signup Failed: Incorrect Input"); return }
    hashedpassword, err := HashPassword(input.Password);
    if err != nil { fmt.Println(err.Error()); c.JSON(http.StatusBadRequest, "Signup Failed: Password unhashable"); return }
    _, err = DB.Exec("insert into users (name, username, password, email) values ($1, $2, $3, $4);", input.Name, input.Username, hashedpassword, input.Email);
    if err != nil { fmt.Println(err.Error()); c.JSON(http.StatusBadRequest, "Signup Failed: Database Rejection"); return }
    c.JSON(http.StatusOK,"Account Created"); 
}




func main() {
    intializeDB();
    defer DB.Close();
	router := gin.Default()
    router.GET("/getusers", getDataExample)
    router.POST("/login", loginEndpoint)
    router.POST("/signup", signupEndpoint)
	router.Run("localhost:8080")

}
