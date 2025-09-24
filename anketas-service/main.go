package main

import (
	"anketas-service/config"
	"anketas-service/infrastructure"
	"anketas-service/service"
	"anketas-service/transport"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {

	err := config.LoadEnv()
	if err != nil {
		log.Println("Произошла проблема при запуске конфига |", err)
		return
	}

	db, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("DATABASE")))
	if err != nil {
		log.Println("Произошла проблема при подключении к БД |", err)
	}

	log.Println("Подключение к БД прошло успешно")

	repo := infrastructure.NewAnketaRepo(db)
	service := service.NewAnketaService(repo)
	handler := transport.NewAnketaHandler(service)

	r := gin.Default()

	handler.RegisterRoutes(r)

	err = r.Run("127.0.0.1:8004")
	if err != nil {
		log.Println("Произошла ошибка при запуске gin |", err)
	}

}
