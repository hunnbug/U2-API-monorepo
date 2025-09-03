package valueObjects

import (
	errs "auth-service/errors"
	"net/mail"
	"regexp"
	"strings"
)

type AuthType int

const (
	EmailAuthType AuthType = iota
	PhoneAuthType
	LoginAuthType
)

func IsEmail(value string) bool {
	return strings.Contains(value, "@")
}

func IsPhoneNumber(value string) bool {
	return strings.Contains(value, "+")
}

func CheckAuthType(value string) AuthType {
	if IsEmail(value) {
		return EmailAuthType
	}

	if IsPhoneNumber(value) {
		return PhoneAuthType
	}

	return LoginAuthType
}

func CreateValueObject(t AuthType, value string) (interface{}, error) {

	switch t {
	case EmailAuthType:
		return NewEmail(value)
	case PhoneAuthType:
		return NewPhone(value)
	case LoginAuthType:
		return NewLogin(value)
	default:
		return nil, errs.ErrWrongAuthType
	}
}

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	if isValidEmail(value) {
		return Email{value}, nil
	}
	return Email{}, errs.ErrInvalidEmail
}

func isValidEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil
}

func (e Email) String() string {
	return e.value
}

type Phone struct {
	value string
}

func NewPhone(value string) (Phone, error) {
	if isValidPhone(value) {
		return Phone{value}, nil
	}
	return Phone{}, errs.ErrInvalidPhone
}

func isValidPhone(value string) bool {
	matched, _ := regexp.MatchString(`^\+7\d{10}$`, value)
	return matched
}

func (p Phone) String() string {
	return p.value
}

// type Password struct {
// 	value string
// }

// func NewPassword(value string) (Password, error) {
// 	if isValidPassword(value) {
// 		passwordHash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
// 		if err != nil {
// 			return Password{""}, err
// 		}
// 		return Password{string(passwordHash)}, nil
// 	}
// 	return Password{}, errs.ErrInvalidPassword
// }

// func isValidPassword(value string) bool {
// 	return len(value) >= 8
// }

// func (p Password) String() string {
// 	return p.value
// }

type Login struct {
	value string
}

func NewLogin(value string) (Login, error) {
	if isValidLogin(value) {
		return Login{value}, nil
	}
	return Login{}, errs.ErrInvalidLogin
}

func isValidLogin(value string) bool {
	if len(value) < 4 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, value)
	return matched
}

func (l Login) String() string {
	return l.value
}
