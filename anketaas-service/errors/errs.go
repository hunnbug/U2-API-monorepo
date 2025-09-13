package errors

import "errors"

//
// ошибки valueOjbects
var ErrInvalidLogin = errors.New("Некорректный юзернейм, он должен быть длинее 4 символов и не содержать специальные знаки")

//
// ошибки доменных объедков
var ErrInvalidGender = errors.New("некорректный гендер")

var ErrInvalidPreferredGender = errors.New("некорректный предпочтительный гендер")

var ErrInvalidTag = errors.New("некорректно указан тег")

var ErrInvalidPhoto = errors.New("некорректная ссылка на фото")

//
// ошибки сервера
var InternalServerError = errors.New("Произошла ошибка на стороне сервера, попробуйте еще раз позже")
