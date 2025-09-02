package domain

import "context"

type UserService interface {
	Auth(ctx context.Context, creds UserCredentials) (token string, err error)
	CreateSession(ctx context.Context, userID string) (token string, err error)
	Logout(ctx context.Context, ID string) error
	ValidateToken(ctx context.Context, token string) error
}
