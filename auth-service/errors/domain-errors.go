package errors

import "errors"

type InvalidEmail error
type InvalidPassword error
type InvalidPhone error
type InvalidLogin error

var ErrInvalidEmail InvalidEmail = errors.New("Невалидный Email адрес!")
var ErrInvalidPassword InvalidPassword = errors.New("Невалидный пароль, он должен быть больше 8 символов!")
var ErrInvalidPhone InvalidPhone = errors.New("Невалидный номер телефона!")
var ErrInvalidLogin InvalidLogin = errors.New("Невалидный логин!")

type LoginAlreadyExists error
type EmailAlreadyExists error
type PhoneAlreadyExists error

var ErrLoginAlreadyExists LoginAlreadyExists = errors.New("Такой логин уже занят!")
var ErrEmailAlreadyExists EmailAlreadyExists = errors.New("Такой адрес электронной почты уже зарегистрирован!")
var ErrPhoneAlreadyExists PhoneAlreadyExists = errors.New("Такой номер телефона уже зарегистрирован!")

type LoginNotExists error
type IncorrectPassword error

var ErrLoginNotExists LoginNotExists = errors.New("Неверный логин!")
var ErrPasswordNotExists LoginNotExists = errors.New("Неверный Пароль!")

type TokenGenerationFailed error

var ErrTokenGenerationFailed TokenGenerationFailed = errors.New("Произошла ошибка при генерации JWT токена")
