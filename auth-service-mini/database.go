package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var redisClient *redis.Client

func initDatabase() error {

	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DATABASE_ADDRESS"),
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := redisClient.Ping(ctx).Err()
	if err != nil {
		return err
	}

	log.Println("Подключено к редису")

	return nil
}

func checkUserCreds(credType string, identifier string, inputPassword string) bool {
	storedHash, err := getPasswordHashFromRedis(credType, identifier)
	if err != nil {
		log.Printf("User not found: %s %s", credType, identifier)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
	if err != nil {
		log.Printf("Неверный пароль")
		return false
	}

	return true
}

func getPasswordHashFromRedis(credType string, identifier string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var redisKey string

	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier
		log.Println("auth:email:", identifier)
	case "login":
		redisKey = "auth:login:" + identifier
		log.Println("auth:login:", identifier)
	case "phone":
		redisKey = "auth:phone:" + identifier
		log.Println("auth:phone:", identifier)
	default:
		return "", fmt.Errorf("invalid credential type: %s", credType)
	}

	storedHash, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return "", fmt.Errorf("not found in Redis: %s", redisKey)
	}

	return storedHash, nil
}

func saveShitToRedis(login, email, phone, password string) {
	ctx := context.Background()

	key := fmt.Sprintf("auth:login:%s", login)
	log.Println("добавляем:", key)
	redisClient.Set(ctx, key, password, 0)

	key = fmt.Sprintf("auth:email:%s", email)
	log.Println("добавляем:", key)
	redisClient.Set(ctx, key, password, 0)

	key = fmt.Sprintf("auth:phone:%s", phone)
	log.Println("добавляем:", key)
	redisClient.Set(ctx, key, password, 0)
}
