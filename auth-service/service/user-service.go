package service

import (
	"auth-service/domain"
	"auth-service/infrastructure"
	"auth-service/valueObjects"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo       infrastructure.RedisUserRepo
	tokenSecret    string
	tokenDuration  time.Duration
	tokenBlacklist map[string]time.Time // Простой in-memory blacklist
}

func NewUserService(
	userRepo infrastructure.RedisUserRepo,
	tokenSecret string,
	tokenDuration time.Duration,
) domain.UserService {
	return &userService{
		userRepo:       userRepo,
		tokenSecret:    tokenSecret,
		tokenDuration:  tokenDuration,
		tokenBlacklist: make(map[string]time.Time),
	}
}

// Auth аутентифицирует пользователя и возвращает токен
func (s *userService) Auth(ctx context.Context, creds domain.UserCredentials) (string, error) {
	// 1. Определяем тип идентификатора
	authType := valueObjects.CheckAuthType(creds.Value)
	identifier, err := valueObjects.CreateValueObject(authType, creds.Value)
	if err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}

	// 2. Ищем хэш пароля в зависимости от типа идентификатора
	var passwordHash string
	switch id := identifier.(type) {
	case valueObjects.Login:
		passwordHash, err = s.userRepo.GetPasswordHashByLogin(id)
	case valueObjects.Email:
		passwordHash, err = s.userRepo.GetPasswordHashByEmail(id)
	case valueObjects.Phone:
		passwordHash, err = s.userRepo.GetPasswordHashByPhone(id)
	default:
		return "", errors.New("unsupported identifier type")
	}

	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// 3. Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	// 4. Получаем userID для создания токена
	var userID string
	switch id := identifier.(type) {
	case valueObjects.Login:
		userID, err = s.userRepo.GetPasswordHashByLogin(id)
	case valueObjects.Email:
		userID, err = s.userRepo.GetPasswordHashByEmail(id)
	case valueObjects.Phone:
		userID, err = s.userRepo.GetPasswordHashByPhone(id)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get user ID: %w", err)
	}

	// 5. Создаем и возвращаем токен
	return s.CreateSession(ctx, userID)
}

// CreateSession создает JWT токен для пользователя
func (s *userService) CreateSession(ctx context.Context, userID string) (string, error) {
	// Создаем claims для JWT
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.tokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Logout добавляет токен в blacklist
func (s *userService) Logout(ctx context.Context, token string) error {
	// Валидируем токен чтобы извлечь expiration time
	claims, err := s.validateTokenAndGetClaims(token)
	if err != nil {
		return err
	}

	// Добавляем токен в blacklist до его естественного истечения
	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	s.tokenBlacklist[token] = expirationTime

	// TODO: Очистка устаревших токенов из blacklist
	// В production лучше использовать Redis для blacklist
	return nil
}

// ValidateToken проверяет валидность токена
func (s *userService) ValidateToken(ctx context.Context, token string) error {
	// 1. Проверяем не в blacklist ли токен
	if expiration, exists := s.tokenBlacklist[token]; exists {
		if time.Now().Before(expiration) {
			return errors.New("token revoked")
		}
		// Удаляем просроченный токен из blacklist
		delete(s.tokenBlacklist, token)
	}

	// 2. Проверяем валидность токена
	_, err := s.validateTokenAndGetClaims(token)
	return err
}

// validateTokenAndGetClaims внутренний метод для валидации и получения claims
func (s *userService) validateTokenAndGetClaims(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.tokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Проверяем не истек ли токен
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().After(time.Unix(int64(exp), 0)) {
			return nil, errors.New("token expired")
		}
	}

	return claims, nil
}

// GetUserIDFromToken извлекает userID из токена (дополнительный полезный метод)
func (s *userService) GetUserIDFromToken(token string) (string, error) {
	claims, err := s.validateTokenAndGetClaims(token)
	if err != nil {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id not found in token")
	}

	return userID, nil
}
