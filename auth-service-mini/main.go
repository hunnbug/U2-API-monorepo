package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	err := LoadEnv()
	if err != nil {
		log.Println("Не удалось загрузить .env файл", err)
		return
	}

	initDatabase()

	router := gin.Default()

	router.POST("/userReg", saveUserRegToRedis)
	router.POST("/login", login)
	router.POST("/verify", verifyToken)
	router.POST("/saveAnketaId", saveAnketaId)
	router.POST("/saveAnketaIdToAll", saveAnketaIdToAll)
	router.POST("/getAnketaId", getAnketaId)
	router.POST("/getAllUserCreds", getAllUserCredsHandler)
	router.POST("/saveUserId", saveUserId)
	router.POST("/saveUserIdToAll", saveUserIdToAll)
	router.POST("/getUserId", getUserId)

	err = router.Run("127.0.0.1:8001")
	if err != nil {
		log.Println("Произошла ошибка при запуске gin:", err)
	}
}
