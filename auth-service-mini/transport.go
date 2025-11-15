package main

import (
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
)

func saveUserRegToRedis(c *gin.Context) {

	var userStrings struct {
		Login    string
		Email    string
		Phone    string
		Password string
		UserId   string `json:"user_id"`
	}

	c.ShouldBindJSON(&userStrings)

	// Сохраняем пароли в Redis
	saveShitToRedis(userStrings.Login, userStrings.Email, userStrings.Phone, userStrings.Password)

	// Сохраняем user_id в Redis по всем кредам
	if userStrings.UserId != "" {
		err := saveUserIdToAllCredTypes(userStrings.Login, userStrings.Email, userStrings.Phone, userStrings.UserId)
		if err != nil {
			log.Printf("Ошибка сохранения user_id в Redis: %v", err)
		}
	}

}

func saveAnketaId(c *gin.Context) {
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
		AnketaId   string `json:"anketa_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	err := saveAnketaIdToRedis(request.CredType, request.Identifier, request.AnketaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID анкеты"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID анкеты успешно сохранен"})
}

func saveAnketaIdToAll(c *gin.Context) {
	var request struct {
		Login    string `json:"login" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
		AnketaId string `json:"anketa_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	err := saveAnketaIdToAllCredTypes(request.Login, request.Email, request.Phone, request.AnketaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID анкеты"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID анкеты успешно сохранен по всем типам учетных данных"})
}

func getAnketaId(c *gin.Context) {
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	anketaId, err := getAnketaIdFromRedis(request.CredType, request.Identifier)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID анкеты не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"anketa_id": anketaId})
}

func getAllUserCredsHandler(c *gin.Context) {
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	login, email, phone, err := getAllUserCreds(request.CredType, request.Identifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения учетных данных"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"login": login,
		"email": email,
		"phone": phone,
	})
}

func saveUserId(c *gin.Context) {
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
		UserId     string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	err := saveUserIdToRedis(request.CredType, request.Identifier, request.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID пользователя успешно сохранен"})
}

func saveUserIdToAll(c *gin.Context) {
	var request struct {
		Login  string `json:"login" binding:"required"`
		Email  string `json:"email" binding:"required"`
		Phone  string `json:"phone" binding:"required"`
		UserId string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	err := saveUserIdToAllCredTypes(request.Login, request.Email, request.Phone, request.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID пользователя успешно сохранен по всем типам учетных данных"})
}

func getUserId(c *gin.Context) {
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	userId, err := getUserIdFromRedis(request.CredType, request.Identifier)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID пользователя не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userId})
}
