package main

import (
	"log"
	"os"
	"user-service/config"
	"user-service/infrastructure"
	"user-service/service"
	"user-service/transport"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {

	err := config.LoadEnv()
	if err != nil {
		log.Println("Произошла ошибка при загрузке переменных окружения")
	}
	db, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("DATABASE")))
	if err != nil {
		log.Println("Произошла ошибка при подключении к базе данных")
	}

	repo := infrastructure.NewMongoRepo(db)
	service := service.NewUserService(repo)
	handler := transport.NewUserHandler(service)

	r := gin.Default()
	handler.RegisterRoutes(r)
}
