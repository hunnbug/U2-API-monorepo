package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client
var messagesCollection *mongo.Collection

func main() {
	// Подключение к MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27019")
	var err error
	client, err = mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Проверка подключения
	err = client.Ping(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	messagesCollection = client.Database("krya").Collection("messages")

	// Настройка роутера
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API endpoints
	router.POST("/send", sendMessage)
	router.GET("/conversation/:senderId/:receiverId", getConversation)
	router.GET("/conversations/:userId", getUserConversations)
	router.PUT("/read/:messageId", markAsRead)

	log.Println("Messages service starting on port 8005...")
	router.Run(":8005")
}
