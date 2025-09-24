package domain

import (
	errs "anketas-service/errors"
	"anketas-service/valueObjects"
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
	case "Asdd", "jsad": // поменять на реальные
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

type Anketa struct {
	ID              uuid.UUID
	Username        valueObjects.Username
	Gender          AnketaGender
	PreferredGender PreferredAnketaGender
	Description     string
	Tags            []Tag
	Photos          []Photo
}
