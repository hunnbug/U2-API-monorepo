package transport

import (
	"log"
	"net/http"
	"user-service/domain"
	errs "user-service/errors"
	"user-service/valueObjects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(service domain.UserService) UserHandler {
	return UserHandler{service}
}

func (h *UserHandler) Register(c *gin.Context) {
	var request struct {
		Login    string `json:"login" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.Register(request.Login, request.Email, request.Phone, request.Password)
	if err != nil {
		switch err {
		case errs.ErrLoginAlreadyExists, errs.ErrEmailAlreadyExists, errs.ErrPhoneAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errs.ErrInvalidEmail, errs.ErrInvalidPhone,
			errs.ErrInvalidPassword, errs.ErrInvalidLogin:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "Пользователь успешно создан"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(request.Login, request.Password)
	if err != nil {
		switch err {
		case errs.ErrLoginNotExists:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errs.ErrPasswordNotExists:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный айди пользователя!"})
		return
	}

	var request struct {
		Login    *string `json:"login,omitempty"`
		Email    *string `json:"email,omitempty"`
		Phone    *string `json:"phone,omitempty"`
		Password *string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var opts []domain.UpdateOption

	if request.Login != nil {
		loginVO, err := valueObjects.NewLogin(*request.Login)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("данные на обновление:", request.Login)
		opts = append(opts, domain.WithLogin(loginVO))
	}

	if request.Email != nil {
		emailVO, err := valueObjects.NewEmail(*request.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("данные на обновление:", request.Email)
		opts = append(opts, domain.WithEmail(emailVO))
	}

	if request.Phone != nil {
		phoneVO, err := valueObjects.NewPhone(*request.Phone)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("данные на обновление:", request.Phone)
		opts = append(opts, domain.WithPhone(phoneVO))
	}

	if request.Password != nil {
		passwordVO, err := valueObjects.NewPassword(*request.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("данные на обновление:", request.Password)
		opts = append(opts, domain.WithPassword(passwordVO))
	}

	err = h.userService.Update(id, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Пользователь успешно обновлен!"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный айди пользователя"})
		return
	}

	err = h.userService.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Пользователь успешно удален!"})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
	router.PUT("/users/:id", h.UpdateUser)
	router.DELETE("/users/:id", h.DeleteUser)
	router.GET("/users/:id", h.GetUser)
}
