package domain

import "github.com/google/uuid"

type UserService interface {
	Register(login, email, phone, password string) error
	Login(login, password string) (string, error)
	Delete(id uuid.UUID) error
	Update(id uuid.UUID, opts ...UpdateOption) error
	GetUserByID(id uuid.UUID) (User, error)
}
