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
	log.Printf("=== ПРОВЕРКА УЧЕТНЫХ ДАННЫХ ===")
	log.Printf("Тип: %s, Идентификатор: %s", credType, identifier)
	
	storedHash, err := getPasswordHashFromRedis(credType, identifier)
	if err != nil {
		log.Printf("User not found: %s %s", credType, identifier)
		return false
	}

	log.Printf("Найден хеш пароля в Redis")
	
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
	if err != nil {
		log.Printf("Неверный пароль")
		return false
	}

	log.Printf("Пароль верный!")
	return true
}

func getPasswordHashFromRedis(credType string, identifier string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var redisKey string

	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier
		log.Printf("Ищем по email: %s", redisKey)
	case "login":
		redisKey = "auth:login:" + identifier
		log.Printf("Ищем по логину: %s", redisKey)
	case "phone":
		redisKey = "auth:phone:" + identifier
		log.Printf("Ищем по телефону: %s", redisKey)
	default:
		return "", fmt.Errorf("invalid credential type: %s", credType)
	}

	storedHash, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		log.Printf("Ошибка получения из Redis: %s - %v", redisKey, err)
		return "", fmt.Errorf("not found in Redis: %s", redisKey)
	}

	log.Printf("Найден хеш в Redis для ключа: %s", redisKey)
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

func saveAnketaIdToRedis(login, anketaId string) error {
	ctx := context.Background()
	
	key := fmt.Sprintf("auth:login:%s:anketa_id", login)
	log.Println("сохраняем ID анкеты:", key, "=", anketaId)
	
	err := redisClient.Set(ctx, key, anketaId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID анкеты: %v", err)
		return err
	}
	
	return nil
}

func getAnketaIdFromRedis(login string) (string, error) {
	ctx := context.Background()
	
	key := fmt.Sprintf("auth:login:%s:anketa_id", login)
	log.Println("получаем ID анкеты:", key)
	
	anketaId, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		log.Printf("ID анкеты не найден: %v", err)
		return "", err
	}
	
	log.Println("найден ID анкеты:", anketaId)
	return anketaId, nil
}