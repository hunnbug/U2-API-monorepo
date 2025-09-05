package main

import "github.com/gin-gonic/gin"

func main() {

	initDatabase()

	router := gin.Default()

	router.POST("/userReg", saveUserRegToRedis)
	router.POST("/login", login)
	router.POST("/verify", verifyToken)
}
