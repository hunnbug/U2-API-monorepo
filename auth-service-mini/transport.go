package main

import "github.com/gin-gonic/gin"

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
