package domain

import (
	"user-service/valueObjects"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID `bson:"id"`
	Login        valueObjects.Login `bson:"login"`
	PasswordHash valueObjects.Password `bson:"password_hash"`
	PhoneNumber  valueObjects.Phone `bson:"phone_number"`
	Email        valueObjects.Email `bson:"email"`
}

func NewUser(login valueObjects.Login, password valueObjects.Password, phone valueObjects.Phone, email valueObjects.Email) User {
	id := uuid.New()
	return User{
		id,
		login,
		password,
		phone,
		email,
	}
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String()), []byte(password))
	return err == nil
}
