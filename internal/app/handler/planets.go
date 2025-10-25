package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"LABS-BMSTU-BACKEND/internal/app/ds"
	"LABS-BMSTU-BACKEND/internal/app/role"

	"LABS-BMSTU-BACKEND/internal/app/myminio"
	"LABS-BMSTU-BACKEND/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
)

// GetPlanets godoc
// @Summary Получить список планет
// @Description Возвращает список всех планет с возможностью поиска по названию
// @Tags planets
// @Accept json
// @Produce json
// @Param query query string false "Поисковый запрос (по названию планеты)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/planet [get]
func (h *Handler) GetPlanets(ctx *gin.Context) {
	var planets []ds.Planets
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		planets, err = h.Repository.GetPlanets()
	} else {
		planets, err = h.Repository.GetPlanetsByTitle(searchQuery)
	}

	if err != nil {
		logrus.Error(err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var systemID uint = 0
	var planetCount int64 = 0

	if rawUserID, ok := ctx.Get("user_id"); ok {
		var userID uint
		switch v := rawUserID.(type) {
		case uint:
			userID = v
		case int:
			userID = uint(v)
		case float64:
			userID = uint(v)
		default:
			logrus.Warnf("GetPlanets: unexpected user_id type: %T", rawUserID)
			userID = 0
		}

		if userID != 0 {
			sysID, err := h.Repository.GetDraftPlanetSystemID(userID)
			if err != nil {
				logrus.Errorf("GetPlanets: error getting draft system id for user %d: %v", userID, err)
			} else {
				systemID = sysID
				if systemID != 0 {
					if cnt, err := h.Repository.GetCountBySystemID(systemID); err != nil {
						logrus.Errorf("GetPlanets: error getting count for system %d: %v", systemID, err)
					} else {
						planetCount = cnt
					}
				}
			}
		}
		logrus.Infof("User %v (system %v) has %d planets", userID, systemID, planetCount)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"planets":      planets,
		"query":        searchQuery,
		"planet_count": planetCount,
		"system_id":    systemID,
	})
}

// GetPlanetById godoc
// @Summary Получить планету по ID
// @Description Возвращает информацию о планете по её идентификатору
// @Tags planets
// @Accept json
// @Produce json
// @Param planet_id path int true "ID планеты"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/planet/{planet_id} [get]
func (h *Handler) GetPlanetById(ctx *gin.Context) {
	planet_id_str := ctx.Param("planet_id")
	planet_id, err := strconv.Atoi(planet_id_str)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet, err := h.Repository.GetPlanet(uint(planet_id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if planet.IsDelete {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("планета не найдена"))
		return
	}

	h.successHandler(ctx, "planet", planet)
}

// AddPlanetToSystem godoc
// @Summary Добавить планету в систему
// @Description Добавляет планету в черновик планетной системы текущего пользователя
// @Tags planets
// @Accept json
// @Produce json
// @Param planet_id path int true "ID планеты"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet/add/{planet_id} [post]
func (h *Handler) AddPlanetToSystem(ctx *gin.Context) {
	userIDVal, okID := ctx.Get("user_id")
	roleVal, okRole := ctx.Get("user_role")
	if !okID || !okRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case float64:
		userID = uint(v)
	default:
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("неверный тип user_id"))
		return
	}

	userRole := roleVal.(role.Role)
	if userRole != role.User {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("только пользователи могут добавлять планеты"))
		return
	}

	planet_id, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	system_id, err := h.Repository.GetDraftPlanetSystemID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if system_id == 0 {
		system_id, err = h.Repository.CreateNewDraftPlanetSystem(userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("не удалось создать черновик системы"))
			return
		}
	}

	system, err := h.Repository.GetPlanetSystemByID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система не найдена"))
		return
	}
	if system.UserID != userID || system.Status != "Черновик" {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("нельзя добавлять планеты в чужую или нечерновую систему"))
		return
	}

	if err := h.Repository.AddPlanetToSystem(uint(planet_id), system.PlanetSystemID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successAddHandler(ctx, "message", "Планета успешно добавлена в систему")
}



type PlanetInput struct {
	PlanetTitle string  `json:"planet_title"`
	Description string  `json:"description"`
	Albedo      float64 `json:"albedo"`
	IsDelete    bool    `json:"is_delete"`
}

// CreatePlanet godoc
// @Summary Создать новую планету
// @Description Создает новую планету (только для модераторов и админов)
// @Tags planets
// @Accept json
// @Produce json
// @Param input body PlanetInput true "Данные для создания планеты"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet/ [post]
func (h *Handler) CreatePlanet(ctx *gin.Context) {
	roleVal, okRole := ctx.Get("user_role")
	if !okRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	userRole := roleVal.(role.Role)
	if userRole != role.Moderator && userRole != role.Admin {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("недостаточно прав"))
		return
	}

	var new_planet PlanetInput
	if err := ctx.ShouldBindJSON(&new_planet); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet := ds.Planets{
		PlanetTitle:       new_planet.PlanetTitle,
		PlanetDescription: new_planet.Description,
		Albedo:            new_planet.Albedo,
		IsDelete:          false,
	}

	planet_id, err := h.Repository.CreatePlanet(planet)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "planet_id", planet_id)
}

// UpdatePlanet godoc
// @Summary Обновить информацию о планете
// @Description Обновляет информацию о планете (только для модераторов и админов)
// @Tags planets
// @Accept json
// @Produce json
// @Param planet_id path int true "ID планеты"
// @Param input body PlanetInput true "Данные для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet/{planet_id} [put]
func (h *Handler) UpdatePlanet(ctx *gin.Context) {
	roleVal, okRole := ctx.Get("user_role")
	if !okRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	userRole := roleVal.(role.Role)
	if userRole != role.Moderator && userRole != role.Admin {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("недостаточно прав"))
		return
	}

	planet_id, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var planet_input PlanetInput
	if err := ctx.ShouldBindJSON(&planet_input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdatePlanet(uint(planet_id), planet_input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Изменено успешно")
}

// DeletePlanet godoc
// @Summary Удалить планету
// @Description Удаляет планету (только для модераторов и админов)
// @Tags planets
// @Accept json
// @Produce json
// @Param planet_id path int true "ID планеты"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet/{planet_id} [delete]
func (h *Handler) DeletePlanet(ctx *gin.Context) {
	roleVal, okRole := ctx.Get("user_role")
	if !okRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	userRole := roleVal.(role.Role)
	if userRole != role.Moderator && userRole != role.Admin {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("недостаточно прав"))
		return
	}

	planet_id, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet, err := h.Repository.GetPlanet(uint(planet_id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("планета не найдена"))
		return
	}

	if planet.PlanetImage != "" {
		objectName := utils.ExtractObjectName(planet.PlanetImage)

		if err := h.Minio.RemoveObject(myminio.BucketName, objectName); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка удаления объекта %s из MinIO: %w", objectName, err))
			return
		}

		if exists, err := h.Minio.StatObject(myminio.BucketName, objectName, minio.StatObjectOptions{}); err == nil && exists.ETag != "" {
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("объект %s не был удалён из MinIO", objectName))
			return
		}
	}

	if err := h.Repository.DeletePlanet(uint(planet_id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Успешно удалена")
}

// AddImage godoc
// @Summary Добавить изображение планеты
// @Description Загружает изображение для планеты
// @Tags planets
// @Accept multipart/form-data
// @Produce json
// @Param planet_id path int true "ID планеты"
// @Param file formData file true "Изображение планеты"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/planet/image/add/{planet_id} [post]
func (h *Handler) AddImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	planet_id := ctx.Param("planet_id")

	if planet_id == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("planet id not found"))
		return
	}
	if header == nil || header.Size == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("header id not found"))
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	defer func(file multipart.File) {
		if cerr := file.Close(); cerr != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, cerr)
			return
		}
	}(file)

	newImageURL, errorCode, errImage := h.createPlanetImage(&file, header, planet_id)
	if errImage != nil {
		h.errorHandler(ctx, errorCode, errImage)
		return
	}

	h.successAddHandler(ctx, "planet_image", newImageURL)
}


// Функция записи фото в минио
func (h *Handler) createPlanetImage(file *multipart.File, header *multipart.FileHeader, planet_id string) (string, int, error) {
	newImageURL, errMinio := h.createImageInMinio(file, header)
	if errMinio != nil {
		return "", http.StatusInternalServerError, errMinio
	}
	if err := h.Repository.UpdatePlanetImage(planet_id, newImageURL); err != nil {
		return "", http.StatusInternalServerError, err
	}
	return newImageURL, 0, nil
}

