package domain

import "github.com/google/uuid"

type UserService interface {
	Register(login, email, phone, password string) (uuid.UUID, error)
	Login(login, password string) (string, error)
	Delete(id uuid.UUID) error
	Update(id uuid.UUID, opts ...UpdateOption) error
	GetUserByID(id uuid.UUID) (User, error)
	CheckLoginExists(login string) (bool, error)
	CheckEmailExists(email string) (bool, error)
	CheckPhoneExists(phone string) (bool, error)
}
