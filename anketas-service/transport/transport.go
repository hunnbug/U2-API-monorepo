package transport

import (
	"anketas-service/domain"
	errs "anketas-service/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnketaHandler struct {
	service domain.AnketaService
}

func NewAnketaHandler(service domain.AnketaService) AnketaHandler {
	return AnketaHandler{service}
}

type CreateAnketaRequest struct {
	Username        string   `json:"username" binding:"required"`
	Gender          string   `json:"gender" binding:"required"`
	PreferredGender string   `json:"preferred_gender" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Tags            []string `json:"tags" binding:"required"`
	Photos          []string `json:"photos" binding:"required"`
}

type UpdateAnketaRequest struct {
	ID              string   `json:"id" binding:"required"`
	Username        string   `json:"username,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	PreferredGender string   `json:"preferred_gender,omitempty"`
	Description     string   `json:"description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Photos          []string `json:"photos,omitempty"`
}

func (h *AnketaHandler) CreateAnketa(c *gin.Context) {
	var req CreateAnketaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.InternalServerError})
		return
	}

	ctx := c.Request.Context()
	err := h.service.Create(
		ctx,
		req.Username,
		req.Gender,
		req.PreferredGender,
		req.Description,
		req.Tags,
		req.Photos,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Анкета успешно создана",
	})
}

func (h *AnketaHandler) GetAnketaByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат ID",
		})
		return
	}

	ctx := c.Request.Context()
	anketa, err := h.service.GetAnketaByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, anketa)
}

func (h *AnketaHandler) UpdateAnketa(c *gin.Context) {
	var req UpdateAnketaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	updateData := make(map[string]interface{})

	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.Gender != "" {
		updateData["gender"] = req.Gender
	}
	if req.PreferredGender != "" {
		updateData["preferred_gender"] = req.PreferredGender
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Tags != nil {
		updateData["tags"] = req.Tags
	}
	if req.Photos != nil {
		updateData["photos"] = req.Photos
	}

	// ID обязательно для обновления
	updateData["id"] = req.ID

	ctx := c.Request.Context()
	err := h.service.Update(ctx, updateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Не удалось обновить анкету"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Анкета успешно обновлена",
	})
}

func (h *AnketaHandler) DeleteAnketa(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат ID",
		})
		return
	}

	ctx := c.Request.Context()
	err = h.service.Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ошибка при удалении анкеты"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Анкета успешно удалена",
	})
}

func (h *AnketaHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/create", h.CreateAnketa)
	r.GET("/anketa/:id", h.GetAnketaByID)
	r.PUT("/anketa/:id", h.UpdateAnketa)
	r.DELETE("/anketa/:id", h.DeleteAnketa)
}
