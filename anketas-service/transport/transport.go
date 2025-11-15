package transport

import (
	"anketas-service/domain"
	"anketas-service/infrastructure"
	errs "anketas-service/errors"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnketaHandler struct {
	service domain.AnketaService
	s3Storage *infrastructure.S3Storage
}

func NewAnketaHandler(service domain.AnketaService, s3Storage *infrastructure.S3Storage) AnketaHandler {
	return AnketaHandler{service: service, s3Storage: s3Storage}
}

type CreateAnketaRequest struct {
	Username        string   `json:"username" binding:"required"`
	Age             int      `json:"age" binding:"required"`
	Gender          string   `json:"gender" binding:"required"`
	PreferredGender string   `json:"preferred_gender" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Tags            []string `json:"tags" binding:"required"`
	Photos          []string `json:"photos" binding:"required"`
	LikedBy         []string `json:"liked_by"`
	CredType        string   `json:"cred_type"`
	Identifier      string   `json:"identifier"`
}

type UpdateAnketaRequest struct {
	Username        string   `json:"username,omitempty"`
	Age             int      `json:"age,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	PreferredGender string   `json:"preferred_gender,omitempty"`
	Description     string   `json:"description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Photos          []string `json:"photos,omitempty"`
	LikedBy         []string `json:"liked_by,omitempty"`
	Action          string   `json:"action,omitempty"`
	CurrentUserAnketaId string `json:"current_user_anketa_id,omitempty"`
}

func (h AnketaHandler) CreateAnketa(c *gin.Context) {
	var req CreateAnketaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.InternalServerError.Error()})
		log.Println("Произошла ошибка при биндинге структуры запроса создания анкеты |", err)
		return
	}

	ctx := c.Request.Context()
	anketaID, err := h.service.Create(
		ctx,
		req.Username,
		req.Age,
		req.Gender,
		req.PreferredGender,
		req.Description,
		req.Tags,
		req.Photos,
		req.LikedBy,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем anketa_id в Redis через auth-service (синхронно, чтобы гарантировать сохранение)
	if req.CredType != "" && req.Identifier != "" {
		log.Printf("Сохраняем anketa_id в Redis: cred_type=%s, identifier=%s, anketa_id=%s", 
			req.CredType, req.Identifier, anketaID.String())
		
		// Сначала сохраняем по переданному типу
		saveAnketaIdToAuthService(req.CredType, req.Identifier, anketaID.String())
		
		// Затем получаем все учетные данные и сохраняем по всем типам
		allCreds, err := getAllUserCredsFromAuthService(req.CredType, req.Identifier)
		if err == nil && allCreds != nil {
			log.Printf("Получены все учетные данные: login=%s, email=%s, phone=%s", 
				allCreds["login"], allCreds["email"], allCreds["phone"])
			saveAnketaIdToAllInAuthService(allCreds["login"], allCreds["email"], allCreds["phone"], anketaID.String())
		} else {
			log.Printf("Не удалось получить все учетные данные, сохранено только по %s", req.CredType)
		}
		
		log.Printf("anketa_id сохранен в Redis (синхронно)")
	} else {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: cred_type или identifier пустые, anketa_id не будет сохранен в Redis")
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: cred_type='%s', identifier='%s'", req.CredType, req.Identifier)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Анкета успешно создана",
		"anketa_id": anketaID.String(),
	})
}

func (h AnketaHandler) GetAnketaByID(c *gin.Context) {
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

func (h AnketaHandler) GetAnketas(c *gin.Context) {
	log.Printf("=== ПОЛУЧЕНИЕ АНКЕТ ДЛЯ МЕТЧИНГА ===")
	
	pref := c.Query("pref")
	id := c.Query("id")
	log.Printf("Запрос: pref=%s, id=%s", pref, id)
	
	preferredGender, err := domain.NewPreferredAnketaGender(pref)
	if err != nil {
		log.Printf("Ошибка при создании PreferredGender из '%s': %v", pref, err)
		c.JSON(500, gin.H{"error": "Произошла ошибка сервера, повторите еще раз позже"})
		return
	}
	log.Printf("PreferredGender создан: %+v", preferredGender)

	parsedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Ошибка при парсинге UUID '%s': %v", id, err)
		c.JSON(500, gin.H{"error": "Произошла ошибка сервера, повторите еще раз позже"})
		return
	}
	log.Printf("UUID распарсен: %s", parsedId.String())

	ctx := c.Request.Context()
	anketas, err := h.service.GetAnketas(ctx, preferredGender, parsedId)
	if err != nil {
		log.Printf("Ошибка получения анкет: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Возвращаем %d анкет клиенту", len(anketas))
	log.Printf("=== КОНЕЦ ПОЛУЧЕНИЯ АНКЕТ ===")
	
	// Убеждаемся, что возвращаем пустой массив, а не null
	if anketas == nil {
		anketas = []domain.Anketa{}
	}
	
	c.JSON(200, gin.H{"anketas": anketas})
}

func (h AnketaHandler) UpdateAnketa(c *gin.Context) {
	log.Printf("=== UpdateAnketa вызван ===")
	log.Printf("Метод запроса: %s", c.Request.Method)
	log.Printf("URL: %s", c.Request.URL.String())
	
	var req UpdateAnketaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("Получен запрос: Action=%s, CurrentUserAnketaId=%s", req.Action, req.CurrentUserAnketaId)

	// Обработка лайка
	if req.Action == "like" && req.CurrentUserAnketaId != "" {
		log.Printf("=== ОБРАБОТКА ЛАЙКА ===")
		idStr := c.Param("id")
		targetAnketaId, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Ошибка парсинга ID целевой анкеты '%s': %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID анкеты"})
			return
		}
		
		// Получаем текущую анкету, чтобы проверить liked_by
		ctx := c.Request.Context()
		anketa, err := h.service.GetAnketaByID(ctx, targetAnketaId)
		if err != nil {
			log.Printf("Ошибка при получении анкеты: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Анкета не найдена"})
			return
		}
		
		// Проверяем, не лайкал ли уже пользователь
		alreadyLiked := false
		for _, likedBy := range anketa.LikedBy {
			if likedBy.String() == req.CurrentUserAnketaId {
				alreadyLiked = true
				break
			}
		}
		
		if alreadyLiked {
			log.Printf("Пользователь уже лайкнул эту анкету")
			c.JSON(http.StatusOK, gin.H{"message": "Лайк уже был поставлен"})
			return
		}
		
		// Добавляем лайк через Update
		updateData := map[string]interface{}{
			"id": idStr,
		}
		
		// Получаем текущий список liked_by и добавляем новый ID
		currentLikedBy := make([]string, 0, len(anketa.LikedBy)+1)
		for _, id := range anketa.LikedBy {
			currentLikedBy = append(currentLikedBy, id.String())
		}
		currentLikedBy = append(currentLikedBy, req.CurrentUserAnketaId)
		updateData["liked_by"] = currentLikedBy
		
		err = h.service.Update(ctx, updateData)
		if err != nil {
			log.Printf("Ошибка при добавлении лайка: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		log.Printf("Лайк успешно добавлен")
		c.JSON(http.StatusOK, gin.H{"message": "Лайк успешно добавлен"})
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
	if req.LikedBy != nil {
		updateData["liked_by"] = req.LikedBy
	}

	updateData["id"] = c.Param("id")

	ctx := c.Request.Context()
	err := h.service.Update(ctx, updateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Не удалось обновить анкету | %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Анкета успешно обновлена",
	})
}

func (h AnketaHandler) DeleteAnketa(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат ID",
		})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ошибка при удалении анкеты"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Анкета успешно удалена",
	})
}

func (h AnketaHandler) GetTags(c *gin.Context) {

	tags := []string{
		"Спорт", "Музыка", "Ранние подъёмы", "Сова",
		"Жаворонок", "Фильмы", "Игры", "Сериалы", "Аниме",
		"Активный отдых", "Рисование", "Путешествия", "Карьера",
		"Книги", "Культурный отдых", "Учёба", "Саморазвитие",
	}

	go func() {
		select {
		case <-c.Request.Context().Done():
			c.JSON(500, gin.H{"error": "Произошла ошибка при получении тегов"})
		default:
			c.JSON(200, gin.H{"tags": tags})
		}
	}()

	c.JSON(200, gin.H{"message": "Начато получение тегов"})
}

// Получение presigned URL для загрузки фотографии
func (h AnketaHandler) GetUploadURL(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(400, gin.H{"error": "user_id обязателен"})
		return
	}

	uploadURL, err := h.s3Storage.GenerateUploadURL(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Ошибка создания presigned URL: %v", err)
		c.JSON(500, gin.H{"error": "Ошибка создания URL для загрузки"})
		return
	}

	c.JSON(200, gin.H{
		"upload_url": uploadURL,
		"message": "URL для загрузки создан",
	})
}

func saveAnketaIdToAuthService(credType, identifier, anketaId string) {
	url := "http://127.0.0.1:8001/saveAnketaId"
	data := map[string]string{
		"cred_type": credType,
		"identifier": identifier,
		"anketa_id": anketaId,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка маршалинга данных для сохранения anketa_id: %v", err)
		return
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ошибка сохранения anketa_id в auth-service: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка сохранения anketa_id: статус код %d", resp.StatusCode)
	} else {
		log.Printf("anketa_id успешно сохранен в auth-service")
	}
}

func getAllUserCredsFromAuthService(credType, identifier string) (map[string]string, error) {
	url := "http://127.0.0.1:8001/getAllUserCreds"
	data := map[string]string{
		"cred_type": credType,
		"identifier": identifier,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("статус код %d", resp.StatusCode)
	}
	
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

func saveAnketaIdToAllInAuthService(login, email, phone, anketaId string) {
	url := "http://127.0.0.1:8001/saveAnketaIdToAll"
	data := map[string]string{
		"login": login,
		"email": email,
		"phone": phone,
		"anketa_id": anketaId,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка маршалинга данных для сохранения anketa_id по всем типам: %v", err)
		return
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ошибка сохранения anketa_id по всем типам в auth-service: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Ошибка сохранения anketa_id по всем типам: статус код %d, тело: %s", resp.StatusCode, string(body))
	} else {
		log.Printf("anketa_id успешно сохранен по всем типам учетных данных в auth-service")
	}
}

func (h AnketaHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/create", h.CreateAnketa)
	r.GET("/anketa/:id", h.GetAnketaByID)
	r.PUT("/anketa/:id", h.UpdateAnketa)
	r.DELETE("/anketa/:id", h.DeleteAnketa)
	r.GET("/anketas/match", h.GetAnketas)
	r.GET("/tags", h.GetTags)
	r.GET("/upload-url", h.GetUploadURL)
}
