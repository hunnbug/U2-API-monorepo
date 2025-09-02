package domain

import (
	"auth-service/valueObjects"
	"time"
)

type UserRepo interface {
	FindPasswordHashByLogin(login valueObjects.Login) (string, error)
	FindPasswordHashByPhone(login valueObjects.Phone) (string, error)
	FindPasswordHashByEmail(login valueObjects.Email) (string, error)
	SaveToken(userID string, refreshToken string, expiresAt time.Time) error
	FindToken(userID string, refreshToken string, expiresAt time.Time) error
	DeleteToken(userID string, refreshToken string) error
	DeleteAllRefreshTokens(userID string) error
}
