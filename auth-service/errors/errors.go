package errors

import "errors"

//
// ошибки пользовательского ввода
type InvalidEmail error
type InvalidPassword error
type InvalidPhone error
type InvalidLogin error

var ErrInvalidEmail InvalidEmail = errors.New("Невалидный Email адрес!")
var ErrInvalidPassword InvalidPassword = errors.New("Невалидный пароль, он должен быть больше 8 символов!")
var ErrInvalidPhone InvalidPhone = errors.New("Невалидный номер телефона!")
var ErrInvalidLogin InvalidLogin = errors.New("Невалидный логин!")

//
// ошибки пользовательского ввода
type LoginNotExists error
type IncorrectPassword error

var ErrLoginNotExists LoginNotExists = errors.New("Неверный логин!")
var ErrPasswordNotExists LoginNotExists = errors.New("Неверный Пароль!")

//
// ошибка при генерации токена
type TokenGenerationFailed error

var ErrTokenGenerationFailed TokenGenerationFailed = errors.New("Произошла ошибка при генерации JWT токена")

//
// ошибки авторизации
type WrongAuthType error

var ErrWrongAuthType WrongAuthType = errors.New("Неверные данные для авторизации")

//
// ошибки редиса
type NotFoundInDB error

var ErrNotFoundInDB NotFoundInDB = errors.New("Не удалось найти данные в базе данных")
