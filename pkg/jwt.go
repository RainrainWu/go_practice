package main

import (

	"fmt"
	"strings"
	"strconv"
	"time"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	gin "github.com/gin-gonic/gin"
)

type Claims struct {

	Account string	`json:"account`
	Role	string 	`json:"role"`
	jwt.StandardClaims
}

type User struct {

	Account		string
	Password	string
}

var (

	jwtSecret = []byte("secret")
	router = gin.Default()
)

func errReport(err error, c *gin.Context) {

	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

func login(c *gin.Context) {

	body := User{}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		errReport(err, c)
		return
	}

	if body.Account == "Rain" && body.Password == "0114" {
		
		now := time.Now()
		jwtId := body.Account + strconv.FormatInt(now.Unix(), 10)
		role := "member"
		claims := Claims {
			Account: 		body.Account,
			Role: 			role,
			StandardClaims:	jwt.StandardClaims{
				Audience:	body.Account,
				ExpiresAt:	now.Add(20 * time.Second).Unix(),
				Id:			jwtId,
				IssuedAt:	now.Unix(),
				Issuer:		"ginJWT",
				NotBefore:	now.Add(10 * time.Second).Unix(),
				Subject: 	body.Account,
			},
		}
		tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token, err := tokenClaims.SignedString(jwtSecret)
		if err != nil {
			errReport(err, c)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"message": "Unauthorized",
	})
}

func profile(c *gin.Context) {
	
	if c.MustGet("account") == "Rain" && c.MustGet("Role") == "Member" {
		c.JSON(http.StatusOK, gin.H{
			"name": "Rain",
			"age": 20,
			"hobby": "music",
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Cannot find the record",
	})
}

func AuthRequired(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	token := strings.Split(auth, "Bearer ")[1]
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecret, nil
	})

	if err != nil {
		var message string
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors & jwt.ValidationErrorMalformed != 0 {
				message = "token is malformed"
			} else if ve.Errors & jwt.ValidationErrorUnverifiable != 0{
				message = "token could not be verified because of signing problems"
			} else if ve.Errors & jwt.ValidationErrorSignatureInvalid != 0 {
				message = "signature validation failed"
			} else if ve.Errors & jwt.ValidationErrorExpired != 0 {
				message = "token is expired"
			} else if ve.Errors & jwt.ValidationErrorNotValidYet != 0 {
				message = "token is not yet valid before sometime"
			} else {
				message = "can not handle this token"
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": message,
		})
		c.Abort()
		return
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		fmt.Println("account:", claims.Account)
		fmt.Println("role:", claims.Role)
		c.Set("account", claims.Account)
		c.Set("role", claims.Role)
		c.Next()
	} else {
		c.Abort()
		return
	}
}

func main() {

	router.POST("/login", login)

	// protected member router
	authorized := router.Group("/")
	authorized.Use(AuthRequired)
	{
		authorized.GET("/member/profile", profile)
	}

	router.Run()
}