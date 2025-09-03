package transport

import (
	"auth-service/domain"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService domain.UserService
}

func NewAuthHandler(userService domain.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

type LoginRequest struct {
	Value    string `json:"value" binding:"required"` // login, email или phone
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LogoutRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/validate", h.ValidateToken)
		authGroup.POST("/session", h.CreateSession)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	creds := domain.UserCredentials{
		Value:    req.Value,
		Password: req.Password,
	}

	ctx, cancel := getContext()
	defer cancel()

	token, err := h.userService.Auth(ctx, creds)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// Logout обрабатывает выход из системы
// @Summary Выход из системы
// @Description Инвалидация токена пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Токен для инвалидации"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	ctx, cancel := getContext()
	defer cancel()
	if err := h.userService.Logout(ctx, req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Successfully logged out"})
}

// ValidateToken проверяет валидность токена
// @Summary Проверка токена
// @Description Проверяет валидность JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ValidateRequest true "Токен для проверки"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	ctx, cancel := getContext()
	defer cancel()

	if err := h.userService.ValidateToken(ctx, req.Token); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Token is valid"})
}

// CreateSession создает новую сессию (токен) для пользователя
// @Summary Создание сессии
// @Description Создает JWT токен для указанного userID
// @Tags auth
// @Accept json
// @Produce json
// @Param userID query string true "User ID"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/session [post]
func (h *AuthHandler) CreateSession(c *gin.Context) {
	userID := c.Query("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "userID is required"})
		return
	}

	ctx, cancel := getContext()
	defer cancel()
	token, err := h.userService.CreateSession(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// Middleware для проверки аутентификации
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid authorization format"})
			c.Abort()
			return
		}

		token := authHeader[7:] // Убираем "Bearer " префикс

		ctx, cancel := getContext()
		defer cancel()

		if err := h.userService.ValidateToken(ctx, token); err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
			c.Abort()
			return
		}

		// Токен валиден, продолжаем обработку
		c.Next()
	}
}

// Response structures
type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// Вспомогательная функция для обработки ошибок аутентификации
func handleAuthError(c *gin.Context, err error) {
	switch err.Error() {
	case "invalid password", "authentication failed":
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	case "unsupported identifier type", "invalid credentials":
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*10)
}
