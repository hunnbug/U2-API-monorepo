package service

import (
	"anketas-service/domain"
	"anketas-service/valueObjects"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type AnketaService struct {
	repo domain.AnketaRepository
}

func NewAnketaService(repo domain.AnketaRepository) AnketaService {
	return AnketaService{repo}
}

var (
	ErrAnketaIDRequired       = errors.New("ID анкеты обязателен для обновления")
	ErrInvalidAnketaID        = errors.New("неверный формат ID анкеты")
	ErrAnketaNotFound         = errors.New("анкета не найдена")
	ErrInvalidGender          = errors.New("неверный пол")
	ErrInvalidPreferredGender = errors.New("неверный предпочитаемый пол")
	ErrInvalidTag             = errors.New("неверный тег")
	ErrInvalidPhoto           = errors.New("неверная ссылка на фото")
	ErrInvalidUsername        = errors.New("неверный юзернейм")
)

func (s AnketaService) Create(
	ctx context.Context,
	username string,
	gender string,
	preferredGender string,
	description string,
	tags []string,
	photos []string,
) error {

	log.Println("Сервис начал создание анкеты")

	usernameVO, err := valueObjects.NewUsername(username)
	if err != nil {
		return fmt.Errorf("неверное имя пользователя: %w", err)
	}

	anketaGender, err := domain.NewAnketaGender(gender)
	if err != nil {
		return fmt.Errorf("неверный пол")
	}

	preferredAnketaGender, err := domain.NewPreferredAnketaGender(preferredGender)
	if err != nil {
		return fmt.Errorf("неверный предпочитаемый пол: %w", err)
	}

	var validatedTags []domain.Tag
	for _, tagValue := range tags {
		tag, err := domain.NewTag(tagValue)
		if err != nil {
			return fmt.Errorf("неверный тег '%s': %w", tagValue, err)
		}
		validatedTags = append(validatedTags, tag)
	}

	var validatedPhotos []domain.Photo
	for _, photoURL := range photos {
		photo, err := domain.NewPhoto(photoURL)
		if err != nil {
			return fmt.Errorf("неверная ссылка на фото '%s': %w", photoURL, err)
		}
		validatedPhotos = append(validatedPhotos, photo)
	}

	anketa := domain.Anketa{
		ID:              uuid.New(),
		Username:        usernameVO,
		Gender:          anketaGender,
		PreferredGender: preferredAnketaGender,
		Description:     description,
		Tags:            validatedTags,
		Photos:          validatedPhotos,
	}

	log.Println("Сервисный слой создал анкету успешно")

	if err := s.repo.Create(ctx, anketa); err != nil {
		return fmt.Errorf("ошибка при создании анкеты: %w", err)
	}

	return nil
}

func (s AnketaService) GetAnketaByID(ctx context.Context, id uuid.UUID) (domain.Anketa, error) {
	anketa, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Anketa{}, fmt.Errorf("ошибка при получении анкеты: %w", err)
	}
	return anketa, nil
}

func (s AnketaService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("ошибка при удалении анкеты: %w", err)
	}
	return nil
}

func (s AnketaService) Update(ctx context.Context, updateData map[string]interface{}) error {

	log.Println("Сервис начал обновление анкеты")

	idValue, exists := updateData["id"]
	if !exists {
		return errors.New("Проблема при получении ID анкеты")
	}

	var id uuid.UUID
	switch v := idValue.(type) {
	case string:
		parsedID, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("Проблема при получении ID анкеты")
		}
		id = parsedID
	case uuid.UUID:
		id = v
	default:
		return fmt.Errorf("Проблема при получении ID анкеты")
	}

	delete(updateData, "id")

	if err := s.validateUpdateData(updateData); err != nil {
		return err
	}

	log.Println("Данные для обновления верны")

	if err := s.repo.Update(ctx, id, updateData); err != nil {
		return err
	}

	log.Println("Сервис завершил обновление анкеты")

	return nil
}

func (s AnketaService) validateUpdateData(updateData map[string]interface{}) error {
	for key, value := range updateData {
		switch key {
		case "username":
			usernameStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("имя пользователя должно быть строкой")
			}
			if _, err := valueObjects.NewUsername(usernameStr); err != nil {
				return fmt.Errorf("неверное имя пользователя: %w", err)
			}

		case "gender":
			genderStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("пол должен быть строкой")
			}
			if _, err := domain.NewAnketaGender(genderStr); err != nil {
				return fmt.Errorf("неверный пол: %w", err)
			}

		case "preferred_gender":
			prefGenderStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("предпочитаемый пол должен быть строкой")
			}
			if _, err := domain.NewPreferredAnketaGender(prefGenderStr); err != nil {
				return fmt.Errorf("неверный предпочитаемый пол: %w", err)
			}

		case "tags":
			tagsSlice, ok := value.([]string)
			if !ok {
				return fmt.Errorf("теги должны быть массивом строк")
			}
			for _, tagValue := range tagsSlice {
				if _, err := domain.NewTag(tagValue); err != nil {
					return fmt.Errorf("неверный тег '%s': %w", tagValue, err)
				}
			}

		case "photos":
			photosSlice, ok := value.([]string)
			if !ok {
				return fmt.Errorf("фото должны быть массивом строк")
			}
			for _, photoURL := range photosSlice {
				if _, err := domain.NewPhoto(photoURL); err != nil {
					return fmt.Errorf("неверная ссылка на фото '%s': %w", photoURL, err)
				}
			}

		case "description":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("описание должно быть строкой")
			}

		default:
			return fmt.Errorf("неизвестное поле '%s' для обновления", key)
		}
	}

	return nil
}

func (s AnketaService) GetAnketas(ctx context.Context, pref domain.PreferredAnketaGender, limit int) ([]domain.Anketa, error) {

	anketas, err := s.repo.GetAnketas(ctx, pref, limit)
	if err != nil {
		return []domain.Anketa{}, err
	}

	return anketas, nil

}
