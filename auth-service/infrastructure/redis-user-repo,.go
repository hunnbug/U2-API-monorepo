package infrastructure

import (
	"auth-service/domain"
	errs "auth-service/errors"
	"auth-service/valueObjects"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUserRepo struct {
	client *redis.Client
}

func NewRedisUserRepo(client *redis.Client) RedisUserRepo {
	return RedisUserRepo{client}
}

func (r *RedisUserRepo) loginKey(login valueObjects.Login) string {
	return fmt.Sprintf("auth:login:%s", login.String())
}

func (r *RedisUserRepo) phoneKey(phone valueObjects.Phone) string {
	return fmt.Sprintf("auth:phone:%s", phone.String())
}

func (r *RedisUserRepo) emailKey(email valueObjects.Email) string {
	return fmt.Sprintf("auth:email:%s", email.String())
}

func (r *RedisUserRepo) userTokensKey(userID string) string {
	return fmt.Sprintf("auth:tokens:%s", userID)
}

func (r *RedisUserRepo) GetPasswordHashByLogin(login valueObjects.Login) (string, error) {
	ctx, cancel := getContext()
	defer cancel()
	key := r.loginKey(login)

	data, err := r.client.HGet(ctx, key, "password_hash").Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errs.ErrNotFoundInDB
		}
		return "", fmt.Errorf("redis get error: %w", err)
	}

	var userData domain.UserCredentials
	if err := json.Unmarshal(data, &userData); err != nil {
		return "", fmt.Errorf("json unmarshal error: %w", err)
	}

	return userData.Password, nil
}

func (r *RedisUserRepo) GetPasswordHashByPhone(phone valueObjects.Phone) (string, error) {
	ctx, cancel := getContext()
	defer cancel()
	key := r.phoneKey(phone)

	data, err := r.client.HGet(ctx, key, "password_hash").Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errs.ErrNotFoundInDB
		}
		return "", fmt.Errorf("redis get error: %w", err)
	}

	var userData domain.UserCredentials
	if err := json.Unmarshal(data, &userData); err != nil {
		return "", fmt.Errorf("json unmarshal error: %w", err)
	}

	return userData.Password, nil
}

func (r *RedisUserRepo) GetPasswordHashByEmail(email valueObjects.Email) (string, error) {
	ctx, cancel := getContext()
	defer cancel()
	key := r.emailKey(email)

	data, err := r.client.HGet(ctx, key, "password_hash").Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errs.ErrNotFoundInDB
		}
		return "", fmt.Errorf("redis get error: %w", err)
	}

	var userData domain.UserCredentials
	if err := json.Unmarshal(data, &userData); err != nil {
		return "", fmt.Errorf("json unmarshal error: %w", err)
	}

	return userData.Password, nil
}

// SaveToken сохраняет refresh token для пользователя
func (r *RedisUserRepo) SaveToken(userID string, refreshToken string, expiresAt time.Time) error {
	ctx, cancel := getContext()
	defer cancel()
	key := r.userTokensKey(userID)

	// Используем Redis Set для хранения токенов
	// Добавляем токен в set и устанавливаем время жизни для всего ключа
	err := r.client.HSet(ctx, key, refreshToken, time.Now().Unix()).Err()
	if err != nil {
		return fmt.Errorf("redis sadd error: %w", err)
	}

	// Устанавливаем время жизни для всего set'а токенов пользователя
	err = r.client.ExpireAt(ctx, key, expiresAt).Err()
	if err != nil {
		return fmt.Errorf("redis expire error: %w", err)
	}

	return nil
}

// FindToken проверяет существование refresh token для пользователя
func (r *RedisUserRepo) FindToken(userID string, refreshToken string) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()
	key := r.userTokensKey(userID)

	// Проверяем, есть ли токен в set'е
	exists, err := r.client.HExists(ctx, key, refreshToken).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, errs.ErrNotFoundInDB
		}
		return false, fmt.Errorf("redis sismember error: %w", err)
	}

	return exists, nil
}

// DeleteToken удаляет конкретный refresh token
func (r *RedisUserRepo) DeleteToken(userID string, refreshToken string) error {
	ctx, cancel := getContext()
	defer cancel()
	key := r.userTokensKey(userID)

	// Удаляем токен из set'а
	err := r.client.HDel(ctx, key, refreshToken).Err()
	if err != nil {
		return fmt.Errorf("redis srem error: %w", err)
	}

	return nil
}

// DeleteAllTokens удаляет все refresh tokens для пользователя
func (r *RedisUserRepo) DeleteAllTokens(userID string) error {
	ctx, cancel := getContext()
	defer cancel()
	key := r.userTokensKey(userID)

	// Просто удаляем весь ключ
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis del error: %w", err)
	}

	return nil
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*10)
}
