package domain

import "github.com/google/uuid"

type UserRepo interface {
	Create(u User) error
	Update(id uuid.UUID, update UserUpdate) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (User, error)
	FindByLogin(login string) (User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByLogin(login string) (bool, error)
	ExistsByPhone(phone string) (bool, error)
}
