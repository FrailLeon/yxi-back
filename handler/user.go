package handle

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keller0/yxi-back/db"
	"github.com/keller0/yxi-back/middleware"
	"golang.org/x/crypto/bcrypt"
)

type login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type register struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Repass   string `form:"repass" json:"repass" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required"`
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Login return a jwt if user info is valid.
func Login(c *gin.Context) {
	var err error
	var loginJSON login
	if err = c.ShouldBindJSON(&loginJSON); err == nil {
		if !checkUserExist(loginJSON.User) {
			// return if username allready exists
			c.JSON(http.StatusBadRequest, gin.H{"error": "user dose not exists"})
			c.Abort()
			return
		}

		var password string
		var id int64
		err = mysql.Db.QueryRow("SELECT id, password FROM user WHERE username=?", loginJSON.User).Scan(
			&id, &password)
		if err != nil {
			log.Fatal(err.Error())
		}

		if checkPasswordHash(loginJSON.Password, password) {
			exp := time.Now().Add(time.Hour * 1).Unix()
			tokenString, err := mid.JwtGenToken(id, loginJSON.User, exp)
			if err != nil {
				c.JSON(500, gin.H{"message": "Could not generate token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": tokenString})

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func CheckUserExist(c *gin.Context) {
	var user string
	if err := c.ShouldBindQuery(user); err != nil {
		if checkUserExist(user) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		} else {
			c.JSON(http.StatusOK, gin.H{"error": ""})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Register use post data to create a user account
func Register(c *gin.Context) {
	var err error
	var registJSON register
	if err = c.ShouldBindJSON(&registJSON); err == nil {
		if checkUserExist(registJSON.User) {
			// return if username allready exists
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			c.Abort()
			return
		}
		if checkEmailExist(registJSON.Email) {
			// return if username allready exists
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
			c.Abort()
			return
		}
		if registJSON.Password != registJSON.Repass {
			// return if password not match
			c.JSON(http.StatusBadRequest, gin.H{"error": "password not match"})
			c.Abort()
			return
		}

		var runToken = randStringRunes(40)
		password, err := hashPassword(registJSON.Password)
		if err != nil {
			log.Fatal(err.Error())
		}
		insUser, err := mysql.Db.Prepare("INSERT INTO user(username, password, email, run_token) values(?,?,?,?)")
		if err != nil {
			log.Fatal(err.Error())
		}

		_, e := insUser.Exec(registJSON.User, password, registJSON.Email, runToken)
		if e != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"error": ""})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func checkUserExist(username string) bool {

	var id int64
	err := mysql.Db.QueryRow("SELECT id FROM user WHERE username=?", username).Scan(&id)
	if err != nil {
		return false
	}
	return id != 0
}

func checkEmailExist(email string) bool {

	var id int64
	err := mysql.Db.QueryRow("SELECT id FROM user WHERE email=?", email).Scan(&id)
	if err != nil {
		return false
	}
	return id != 0
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}