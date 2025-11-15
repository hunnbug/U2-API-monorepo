package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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

func saveAnketaIdToRedis(credType, identifier, anketaId string) error {
	ctx := context.Background()
	
	var redisKey string
	
	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier + ":anketa_id"
		log.Printf("Сохраняем ID анкеты по email: %s", redisKey)
	case "login":
		redisKey = "auth:login:" + identifier + ":anketa_id"
		log.Printf("Сохраняем ID анкеты по логину: %s", redisKey)
	case "phone":
		redisKey = "auth:phone:" + identifier + ":anketa_id"
		log.Printf("Сохраняем ID анкеты по телефону: %s", redisKey)
	default:
		return fmt.Errorf("invalid credential type: %s", credType)
	}
	
	err := redisClient.Set(ctx, redisKey, anketaId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID анкеты: %v", err)
		return err
	}
	
	log.Printf("ID анкеты успешно сохранен: %s = %s", redisKey, anketaId)
	return nil
}

func saveAnketaIdToAllCredTypes(login, email, phone, anketaId string) error {
	ctx := context.Background()
	
	// Сохраняем по логину
	key := fmt.Sprintf("auth:login:%s:anketa_id", login)
	log.Printf("Сохраняем ID анкеты по логину: %s = %s", key, anketaId)
	err := redisClient.Set(ctx, key, anketaId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID анкеты по логину: %v", err)
		return err
	}
	
	// Сохраняем по email
	key = fmt.Sprintf("auth:email:%s:anketa_id", email)
	log.Printf("Сохраняем ID анкеты по email: %s = %s", key, anketaId)
	err = redisClient.Set(ctx, key, anketaId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID анкеты по email: %v", err)
		return err
	}
	
	// Сохраняем по телефону
	key = fmt.Sprintf("auth:phone:%s:anketa_id", phone)
	log.Printf("Сохраняем ID анкеты по телефону: %s = %s", key, anketaId)
	err = redisClient.Set(ctx, key, anketaId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID анкеты по телефону: %v", err)
		return err
	}
	
	log.Printf("ID анкеты успешно сохранен по всем типам учетных данных")
	return nil
}

func getAnketaIdFromRedis(credType, identifier string) (string, error) {
	ctx := context.Background()
	
	var redisKey string
	
	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier + ":anketa_id"
		log.Printf("Ищем ID анкеты по email: %s", redisKey)
	case "login":
		redisKey = "auth:login:" + identifier + ":anketa_id"
		log.Printf("Ищем ID анкеты по логину: %s", redisKey)
	case "phone":
		redisKey = "auth:phone:" + identifier + ":anketa_id"
		log.Printf("Ищем ID анкеты по телефону: %s", redisKey)
	default:
		return "", fmt.Errorf("invalid credential type: %s", credType)
	}
	
	anketaId, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		log.Printf("ID анкеты не найден для %s %s: %v", credType, identifier, err)
		return "", err
	}
	
	log.Printf("Найден ID анкеты: %s = %s", redisKey, anketaId)
	return anketaId, nil
}

func getAllUserCreds(credType, identifier string) (login, email, phone string, err error) {
	ctx := context.Background()
	
	// 1. Получаем пароль по переданным данным
	password, err := getPasswordHashFromRedis(credType, identifier)
	if err != nil {
		log.Printf("Ошибка получения пароля для %s %s: %v", credType, identifier, err)
		return "", "", "", err
	}
	
	log.Printf("Найден пароль для %s %s, ищем связанные учетные данные", credType, identifier)
	
	// 2. Ищем все ключи с этим паролем
	patterns := []string{"auth:login:*", "auth:email:*", "auth:phone:*"}
	
	for _, pattern := range patterns {
		keys, err := redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			log.Printf("Ошибка поиска ключей по паттерну %s: %v", pattern, err)
			continue
		}
		
		for _, key := range keys {
			storedPassword, err := redisClient.Get(ctx, key).Result()
			if err != nil {
				log.Printf("Ошибка получения значения для ключа %s: %v", key, err)
				continue
			}
			
			if storedPassword == password {
				// Извлекаем тип и значение
				parts := strings.Split(key, ":")
				if len(parts) == 3 {
					credType := parts[1]
					credValue := parts[2]
					
					log.Printf("Найдено совпадение: %s = %s", key, credValue)
					
					switch credType {
					case "login":
						login = credValue
					case "email":
						email = credValue
					case "phone":
						phone = credValue
					}
				}
			}
		}
	}
	
	log.Printf("Результат поиска: login=%s, email=%s, phone=%s", login, email, phone)
	return login, email, phone, nil
}

// Сохранение user_id в Redis по одному типу учетных данных
func saveUserIdToRedis(credType, identifier, userId string) error {
	ctx := context.Background()
	
	var redisKey string
	
	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier + ":user_id"
		log.Printf("Сохраняем ID пользователя по email: %s", redisKey)
	case "login":
		redisKey = "auth:login:" + identifier + ":user_id"
		log.Printf("Сохраняем ID пользователя по логину: %s", redisKey)
	case "phone":
		redisKey = "auth:phone:" + identifier + ":user_id"
		log.Printf("Сохраняем ID пользователя по телефону: %s", redisKey)
	default:
		return fmt.Errorf("invalid credential type: %s", credType)
	}
	
	err := redisClient.Set(ctx, redisKey, userId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID пользователя: %v", err)
		return err
	}
	
	log.Printf("ID пользователя успешно сохранен: %s = %s", redisKey, userId)
	return nil
}

// Сохранение user_id в Redis по всем типам учетных данных
func saveUserIdToAllCredTypes(login, email, phone, userId string) error {
	ctx := context.Background()
	
	// Сохраняем по логину
	key := fmt.Sprintf("auth:login:%s:user_id", login)
	log.Printf("Сохраняем ID пользователя по логину: %s = %s", key, userId)
	err := redisClient.Set(ctx, key, userId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID пользователя по логину: %v", err)
		return err
	}
	
	// Сохраняем по email
	key = fmt.Sprintf("auth:email:%s:user_id", email)
	log.Printf("Сохраняем ID пользователя по email: %s = %s", key, userId)
	err = redisClient.Set(ctx, key, userId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID пользователя по email: %v", err)
		return err
	}
	
	// Сохраняем по телефону
	key = fmt.Sprintf("auth:phone:%s:user_id", phone)
	log.Printf("Сохраняем ID пользователя по телефону: %s = %s", key, userId)
	err = redisClient.Set(ctx, key, userId, 0).Err()
	if err != nil {
		log.Printf("Ошибка сохранения ID пользователя по телефону: %v", err)
		return err
	}
	
	log.Printf("ID пользователя успешно сохранен по всем типам учетных данных")
	return nil
}

// Получение user_id из Redis
func getUserIdFromRedis(credType, identifier string) (string, error) {
	ctx := context.Background()
	
	var redisKey string
	
	switch credType {
	case "email":
		redisKey = "auth:email:" + identifier + ":user_id"
		log.Printf("Ищем ID пользователя по email: %s", redisKey)
	case "login":
		redisKey = "auth:login:" + identifier + ":user_id"
		log.Printf("Ищем ID пользователя по логину: %s", redisKey)
	case "phone":
		redisKey = "auth:phone:" + identifier + ":user_id"
		log.Printf("Ищем ID пользователя по телефону: %s", redisKey)
	default:
		return "", fmt.Errorf("invalid credential type: %s", credType)
	}
	
	userId, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		log.Printf("ID пользователя не найден для %s %s: %v", credType, identifier, err)
		return "", err
	}
	
	log.Printf("Найден ID пользователя: %s = %s", redisKey, userId)
	return userId, nil
}