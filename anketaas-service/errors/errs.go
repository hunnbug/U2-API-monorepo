package errors

import "errors"

//
// ошибки valueOjbects
type InvalidLogin error

var ErrInvalidLogin = errors.New("Некорректный юзернейм, он должен быть длинее 4 символов и не содержать специальные знаки")

//
// ошибки доменных объедков
type InvalidGender error

var ErrInvalidGender = errors.New("некорректный гендер")

type InvalidPreferredGender error

var ErrInvalidPreferredGender = errors.New("некорректный предпочтительный гендер")

type InvalidTag error

var ErrInvalidTag = errors.New("некорректно указан тег")

type InvalidPhoto error

var ErrInvalidPhoto = errors.New("некорректная ссылка на фото")
