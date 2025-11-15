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

	if err := c.ShouldBindJSON(&userStrings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Сохраняем пароли в Redis
	err := saveShitToRedis(userStrings.Login, userStrings.Email, userStrings.Phone, userStrings.Password)
	if err != nil {
		log.Printf("Ошибка сохранения паролей в Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения данных"})
		return
	}

	// Сохраняем user_id в Redis по всем кредам
	if userStrings.UserId != "" {
		err := saveUserIdToAllCredTypes(userStrings.Login, userStrings.Email, userStrings.Phone, userStrings.UserId)
		if err != nil {
			log.Printf("Ошибка сохранения user_id в Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения user_id"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Данные успешно сохранены"})
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
	log.Printf("=== СОХРАНЕНИЕ ANKETA_ID ДЛЯ ВСЕХ ТИПОВ ===")
	
	var request struct {
		Login    string `json:"login" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AnketaId string `json:"anketa_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Ошибка парсинга запроса saveAnketaIdToAll: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	log.Printf("Получен запрос на сохранение anketa_id: login=%s, email=%s, phone=%s, anketa_id=%s", 
		request.Login, request.Email, request.Phone, request.AnketaId)

	if request.Login == "" {
		log.Printf("ОШИБКА: login пустой!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "login обязателен"})
		return
	}

	err := saveAnketaIdToAllCredTypes(request.Login, request.Email, request.Phone, request.AnketaId)
	if err != nil {
		log.Printf("Ошибка сохранения anketa_id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID анкеты"})
		return
	}

	log.Printf("anketa_id успешно сохранен для всех типов учетных данных")
	log.Printf("=== СОХРАНЕНИЕ ЗАВЕРШЕНО ===")
	c.JSON(http.StatusOK, gin.H{"message": "ID анкеты успешно сохранен по всем типам учетных данных"})
}

func getAnketaId(c *gin.Context) {
	log.Printf("=== ПОЛУЧЕНИЕ ANKETA_ID ===")
	
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Ошибка парсинга запроса getAnketaId: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	log.Printf("Запрос anketa_id для %s: %s", request.CredType, request.Identifier)

	anketaId, err := getAnketaIdFromRedis(request.CredType, request.Identifier)
	if err != nil {
		log.Printf("anketa_id не найден для %s: %s, ошибка: %v", request.CredType, request.Identifier, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "ID анкеты не найден"})
		return
	}

	log.Printf("anketa_id найден: %s", anketaId)
	log.Printf("=== ANKETA_ID УСПЕШНО ПОЛУЧЕН ===")
	c.JSON(http.StatusOK, gin.H{"anketa_id": anketaId})
}

func getAllUserCredsHandler(c *gin.Context) {
	log.Printf("=== ПОЛУЧЕНИЕ ВСЕХ УЧЕТНЫХ ДАННЫХ ===")
	
	var request struct {
		CredType   string `json:"cred_type" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Ошибка парсинга запроса getAllUserCreds: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	log.Printf("Запрос всех учетных данных для %s: %s", request.CredType, request.Identifier)

	login, email, phone, err := getAllUserCreds(request.CredType, request.Identifier)
	if err != nil {
		log.Printf("Ошибка получения всех учетных данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения учетных данных"})
		return
	}

	log.Printf("Получены все учетные данные: login=%s, email=%s, phone=%s", login, email, phone)
	log.Printf("=== УЧЕТНЫЕ ДАННЫЕ УСПЕШНО ПОЛУЧЕНЫ ===")

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
