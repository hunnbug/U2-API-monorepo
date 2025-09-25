package domain

import (
	errs "anketas-service/errors"
	"anketas-service/valueObjects"
	"errors"
	"strings"

	"github.com/google/uuid"
)

const (
	Man   = "Мужчина"
	Woman = "Женщина"
)

type AnketaGender struct {
	Value string
}

func NewAnketaGender(value string) (AnketaGender, error) {
	switch value {
	case Man, Woman:
		return AnketaGender{value}, nil
	default:
		return AnketaGender{}, errs.ErrInvalidGender
	}
}

const (
	PreferredWoman = "Женщин"
	PreferredMan   = "Мужчин"
	PreferredBoth  = "Всех"
)

type PreferredAnketaGender struct {
	Value string
}

func NewPreferredAnketaGender(value string) (PreferredAnketaGender, error) {
	switch value {
	case PreferredMan, PreferredWoman, PreferredBoth:
		return PreferredAnketaGender{value}, nil
	default:
		return PreferredAnketaGender{}, errs.ErrInvalidPreferredGender
	}
}

type Tag struct {
	Value string
}

func NewTag(value string) (Tag, error) {
	switch value {
	case "Спорт", "Музыка", "Ранние подъёмы", "Сова",
		"Жаворонок", "Фильмы", "Игры", "Сериалы", "Аниме",
		"Активный отдых", "Рисование", "Путешествия", "Карьера",
		"Книги", "Культурный отдых", "Учёба", "Саморазвитие":
		return Tag{value}, nil
	default:
		return Tag{}, errs.ErrInvalidTag
	}
}

type Photo struct {
	Url string
}

func NewPhoto(url string) (Photo, error) {
	if strings.Contains(url, "https://") {
		return Photo{url}, nil
	} else {
		return Photo{}, errs.ErrInvalidPhoto
	}
}

type Age int

func NewAge(value int) (Age, error) {
	if value <= 0 {
		return 0, errors.New("Возраст не может быть меньше нуля")
	}
	if value > 119 {
		return 0, errors.New("Возраст не может быть больше 119")
	}
	return Age(value), nil
}

func (a Age) Int() int {
	return int(a)
}

type Anketa struct {
	ID              uuid.UUID
	Username        valueObjects.Username
	Age             Age
	Gender          AnketaGender
	PreferredGender PreferredAnketaGender
	Description     string
	Tags            []Tag
	Photos          []Photo
}
