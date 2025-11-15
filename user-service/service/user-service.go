package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"user-service/domain"
	errs "user-service/errors"
	"user-service/valueObjects"

	"github.com/google/uuid"
)

type UserServiceImpl struct {
	repo domain.UserRepo
}

func NewUserService(repo domain.UserRepo) domain.UserService {
	return UserServiceImpl{repo}
}

func (s UserServiceImpl) Register(login, email, phone, password string) (uuid.UUID, error) {
	loginVO, err := valueObjects.NewLogin(login)
	if err != nil {
		return uuid.Nil, err
	}

	emailVO, err := valueObjects.NewEmail(email)
	if err != nil {
		return uuid.Nil, err
	}

	phoneVO, err := valueObjects.NewPhone(phone)
	if err != nil {
		return uuid.Nil, err
	}

	passwordVO, err := valueObjects.NewPassword(password)
	if err != nil {
		return uuid.Nil, err
	}

	if exists, _ := s.repo.ExistsByLogin(loginVO.String()); exists {
		return uuid.Nil, errs.ErrLoginAlreadyExists
	}

	if exists, _ := s.repo.ExistsByEmail(emailVO.String()); exists {
		return uuid.Nil, errs.ErrEmailAlreadyExists
	}

	if exists, _ := s.repo.ExistsByPhone(phoneVO.String()); exists {
		return uuid.Nil, errs.ErrPhoneAlreadyExists
	}

	user := domain.NewUser(loginVO, passwordVO, phoneVO, emailVO)
	var userStrings struct {
		Login    string
		Email    string
		Phone    string
		Password string
		UserId   string `json:"user_id"`
	}
	userStrings.Email = user.Email.String()
	userStrings.Phone = user.PhoneNumber.String()
	userStrings.Login = user.Login.String()
	userStrings.Password = user.PasswordHash.String()
	userStrings.UserId = user.ID.String()

	dataToSend, err := json.Marshal(userStrings)
	if err != nil {
		return uuid.Nil, err
	}

	http.Post("http://localhost:8001/userReg", "application/json", bytes.NewBuffer(dataToSend))

	err = s.repo.Create(user)
	if err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (s UserServiceImpl) Login(login, password string) (string, error) {
	user, err := s.repo.FindByLogin(login)
	if err != nil {
		return "", errs.ErrLoginNotExists
	}

	if !user.CheckPassword(password) {
		return "", errs.ErrPasswordNotExists
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", errs.ErrTokenGenerationFailed
	}

	return token, nil
}

func (s UserServiceImpl) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s UserServiceImpl) Update(id uuid.UUID, opts ...domain.UpdateOption) error {

	update := domain.NewUserUpdate()

	for _, opt := range opts {
		opt(update)
	}

	return s.repo.Update(id, *update)
}

func (s UserServiceImpl) GetUserByID(id uuid.UUID) (domain.User, error) {
	return s.repo.FindByID(id)
}

func (s UserServiceImpl) CheckLoginExists(login string) (bool, error) {
	return s.repo.ExistsByLogin(login)
}

func (s UserServiceImpl) CheckEmailExists(email string) (bool, error) {
	return s.repo.ExistsByEmail(email)
}

func (s UserServiceImpl) CheckPhoneExists(phone string) (bool, error) {
	return s.repo.ExistsByPhone(phone)
}

func (s UserServiceImpl) generateToken(id uuid.UUID) (string, error) {
	return "generated-token-" + id.String(), nil
}
