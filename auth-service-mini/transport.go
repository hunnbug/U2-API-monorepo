package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func saveUserRegToRedis(c *gin.Context) {

	var userStrings struct {
		Login    string
		Email    string
		Phone    string
		Password string
	}

	c.ShouldBindJSON(&userStrings)

	saveShitToRedis(userStrings.Login, userStrings.Email, userStrings.Phone, userStrings.Password)

}

func saveAnketaId(c *gin.Context) {
	var request struct {
		Login    string `json:"login" binding:"required"`
		AnketaId string `json:"anketa_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	err := saveAnketaIdToRedis(request.Login, request.AnketaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ID анкеты"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID анкеты успешно сохранен"})
}

func getAnketaId(c *gin.Context) {
	login := c.Param("login")
	if login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин не указан"})
		return
	}

	anketaId, err := getAnketaIdFromRedis(login)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID анкеты не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"anketa_id": anketaId})
}
