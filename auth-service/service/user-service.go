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

func (s *userService) Auth(ctx context.Context, creds domain.UserCredentials) (string, error) {
	authType := valueObjects.CheckAuthType(creds.Value)
	identifier, err := valueObjects.CreateValueObject(authType, creds.Value)
	if err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}

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

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password))
	if err != nil {
		return "", errors.New("invalid password")
	}

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

	return s.CreateSession(ctx, userID)
}

func (s *userService) CreateSession(ctx context.Context, userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.tokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *userService) Logout(ctx context.Context, token string) error {
	claims, err := s.validateTokenAndGetClaims(token)
	if err != nil {
		return err
	}

	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	s.tokenBlacklist[token] = expirationTime

	return nil
}

func (s *userService) ValidateToken(ctx context.Context, token string) error {
	if expiration, exists := s.tokenBlacklist[token]; exists {
		if time.Now().Before(expiration) {
			return errors.New("token revoked")
		}
		delete(s.tokenBlacklist, token)
	}

	_, err := s.validateTokenAndGetClaims(token)
	return err
}

func (s *userService) validateTokenAndGetClaims(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
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

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().After(time.Unix(int64(exp), 0)) {
			return nil, errors.New("token expired")
		}
	}

	return claims, nil
}

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
