package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("pidorok-key")

func login(c *gin.Context) {

	var request struct {
		Creds    string
		Value    string
		Password string
	}

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных для входа"})
		return
	}

	if !checkUserCreds(request.Creds, request.Value, request.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные для входа"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(40 * time.Second).Unix(), // поменять время истечения токена
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на стороне сервера"})
		return
	}

	// Получаем user_id из Redis
	log.Printf("Пытаемся получить user_id для %s: %s", request.Creds, request.Value)
	userId, err := getUserIdFromRedis(request.Creds, request.Value)
	if err != nil {
		log.Printf("Ошибка получения user_id: %v", err)
		// Не блокируем авторизацию, просто не возвращаем user_id
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
		return
	}
	
	log.Printf("Найден user_id: %s", userId)

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user_id": userId,
	})
}

func authMiddleware(c *gin.Context) {

	authHeader := c.GetHeader("AuthHeader")

	log.Println("заголовок запроса:", authHeader)
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка, проблема с авторизацией"})
		c.Abort()
		return
	}

	tokenString := authHeader[7:]
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		log.Println("Произошла проблема при валидации токена |", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Проблема с авторизацией"})
		c.Abort()
		return
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(40 * time.Second).Unix(), // поменять время истечения токена
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на стороне сервера"})
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString, "status": "Токен верен"})
	}

	c.Next()
}

func verifyToken(c *gin.Context) {

	authMiddleware(c)
	if c.IsAborted() {
		return
	}
}
