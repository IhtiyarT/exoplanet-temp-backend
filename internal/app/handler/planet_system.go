package handler

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"LABS-BMSTU-BACKEND/internal/app/role"
	"LABS-BMSTU-BACKEND/internal/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"LABS-BMSTU-BACKEND/internal/app/dto"
)

// DeletePlanetSystem godoc
// @Summary Удалить планетную систему
// @Description Удаляет черновик планетной системы текущего пользователя
// @Tags planet-system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/delete [delete]
func (h *Handler) DeletePlanetSystem(ctx *gin.Context) {
	currentUserIDVal, ok := ctx.Get("user_id")
	if !ok {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)

	system_id, err := h.Repository.GetDraftPlanetSystemID(currentUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if system_id == 0 {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система для удаления не найдена"))
		return
	}

	sys, err := h.Repository.GetPlanetSystemByID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if sys.UserID != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("только создатель может удалить черновик"))
		return
	}

	if err := h.Repository.DeletePlanetSystem(uint(system_id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "успешно удалено")
}

// GetPlanetSystemByID godoc
// @Summary Получить планетную систему по ID
// @Description Возвращает информацию о планетной системе по её идентификатору
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/{system_id} [get]
func (h *Handler) GetPlanetSystemByID(ctx *gin.Context) {
	currentUserIDVal, hasUser := ctx.Get("user_id")
	currentUserRoleVal, hasRole := ctx.Get("user_role")

	if !hasUser || !hasRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)
	currentUserRole := currentUserRoleVal.(role.Role)

	idStr := ctx.Param("system_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet_system, err := h.Repository.GetPlanetSystemByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if planet_system.Status == "Удалена" {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система не найдена"))
		return
	}

	if currentUserRole == role.User && planet_system.UserID != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("доступ запрещен"))
		return
	}

	h.successHandler(ctx, "planet_system", planet_system)
}

// GetPlanetSystemDraftID godoc
// @Summary Получить ID черновика планетной системы
// @Description Возвращает информацию о черновике планетной системы текущего пользователя
// @Tags planet-system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/draft/id [get]
func (h *Handler) GetPlanetSystemDraftID(ctx *gin.Context) {
	currentUserIDVal, ok := ctx.Get("user_id")
	if !ok {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)

	system_id, err := h.Repository.GetDraftPlanetSystemID(currentUserID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if system_id == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"system_id":       0,
			"planet_count":    0,
			"star_type":       "",
			"star_name":       "",
			"star_luminosity": 0,
			"planets":         []ds.Planets{},
		})
		return
	}

	planet_count, err := h.Repository.GetCountBySystemID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	planet_system, err := h.Repository.GetPlanetSystemAndPlanetsByID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"system_id":       system_id,
		"planet_count":    planet_count,
		"star_type":       planet_system.StarType,
		"star_name":       planet_system.StarName,
		"star_luminosity": planet_system.StarLuminosity,
		"planets":         planet_system.Planets,
	})
}

// GetPlanetSystemsList godoc
// @Summary Получить список планетных систем
// @Description Получить список всех планетных систем (требуется авторизация)
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_status query string false "Статус системы"
// @Param start_date query string false "Начальная дата (YYYY-MM-DD)"
// @Param end_date query string false "Конечная дата (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/list [get]
func (h *Handler) GetPlanetSystemsList(ctx *gin.Context) {
	currentUserIDVal, hasUser := ctx.Get("user_id")
	currentUserRoleVal, hasRole := ctx.Get("user_role")

	if !hasUser || !hasRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)
	currentUserRole := currentUserRoleVal.(role.Role)

	if system_id := ctx.Query("system_id"); system_id != "" {
		system_id_int, err := strconv.Atoi(system_id)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		system, err := h.Repository.GetPlanetSystemByID(uint(system_id_int))
		if err != nil {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}

		if currentUserRole != role.Moderator && currentUserRole != role.Admin && system.UserID != currentUserID {
			h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("доступ запрещен"))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"planet_system": system,
		})
		return
	}

	system_status := ctx.Query("system_status")
	start_date_str := ctx.Query("start_date")
	end_date_str := ctx.Query("end_date")

	var start_date, end_date time.Time
	var err_start, err_end error

	if start_date_str == "" {
		start_date = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		start_date, err_start = utils.ParseDateString(start_date_str)
	}

	if end_date_str == "" {
		end_date = time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
	} else {
		end_date, err_end = utils.ParseDateString(end_date_str)
	}

	if err_end != nil || err_start != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный start_date или end_date"))
		return
	}

	var systems []dto.SystemListItem
	var errRepo error

	if currentUserRole == role.Moderator || currentUserRole == role.Admin {
		systems, errRepo = h.Repository.GetPlanetSystemsForList(system_status, start_date, end_date)
	} else {
		systems, errRepo = h.Repository.GetPlanetSystemsForListByUser(currentUserID, system_status, start_date, end_date)
	}

	if errRepo != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, errRepo)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"planet_systems": systems,
	})
}

// GetPlanetSystemAndPlanetsByID godoc
// @Summary Получить планетную систему с планетами по ID
// @Description Возвращает полную информацию о планетной системе включая список планет
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/{system_id}/planets [get]
func (h *Handler) GetPlanetSystemAndPlanetsByID(ctx *gin.Context) {
	system_id := ctx.Param("system_id")
	id, err := strconv.Atoi(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUserIDVal, okID := ctx.Get("user_id")
	currentUserRoleVal, okRole := ctx.Get("user_role")
	if !okID || !okRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)
	currentUserRole := currentUserRoleVal.(role.Role)

	planet_system, err := h.Repository.GetPlanetSystemAndPlanetsByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if planet_system.Status == "Удалена" {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система не найдена"))
		return
	}

	if currentUserRole == role.User && planet_system.UserID != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("доступ запрещен"))
		return
	}

	planet_count, err := h.Repository.GetCountBySystemID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"planets":         planet_system.Planets,
		"star_type":       planet_system.StarType,
		"star_name":       planet_system.StarName,
		"star_luminocity": planet_system.StarLuminosity,
		"planet_count":    planet_count,
	})
}

type PlanetSystemInput struct {
	StarType       string `json:"star_type"`
	StarName       string `json:"star_name"`
	StarLuminosity uint   `json:"star_luminosity"`
}

// UpdatePlanetSystem godoc
// @Summary Обновить планетную систему
// @Description Обновляет информацию о планетной системе (только для черновиков)
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Param input body PlanetSystemInput true "Данные для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/{system_id} [put]
func (h *Handler) UpdatePlanetSystem(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUserIDVal, hasUser := ctx.Get("user_id")
	currentUserRoleVal, hasRole := ctx.Get("user_role")
	if !hasUser || !hasRole {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)
	currentUserRole := currentUserRoleVal.(role.Role)

	if currentUserRole != role.User {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("только пользователь может изменять систему"))
		return
	}

	system, err := h.Repository.GetPlanetSystemByID(uint(system_id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if system.UserID != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("можно изменять только собственную систему"))
		return
	}

	if system.Status != "Черновик" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("изменять можно только систему в статусе 'Черновик'"))
		return
	}

	var input PlanetSystemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdatePlanetSystem(uint(system_id), input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.successAddHandler(ctx, "message", "обновлено успешно")
}

// SetPlanetSystemFormed godoc
// @Summary Сформировать заявку
// @Description Переводит черновик планетной системы в статус "Сформирована"
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/{system_id}/form [put]
func (h *Handler) SetPlanetSystemFormed(ctx *gin.Context) {
	system_id, err := strconv.ParseUint(ctx.Param("system_id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	id := uint(system_id)

	currentUserIDVal, ok1 := ctx.Get("user_id")
	currentUserRoleVal, ok2 := ctx.Get("user_role")
	if !ok1 || !ok2 {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserID := currentUserIDVal.(uint)
	currentUserRole := currentUserRoleVal.(role.Role)

	if currentUserRole != role.User {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("только пользователь может формировать заявку"))
		return
	}

	system, err := h.Repository.GetPlanetSystemByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if system.UserID != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("можно формировать только собственную заявку"))
		return
	}

	if system.Status != "Черновик" || system.DateCreated.IsZero() {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("заявка должна иметь статус 'Черновик' и дату создания"))
		return
	}

	if err := h.Repository.SetPlanetSystemFormed(id, currentUserID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "заявка успешно сформирована",
	})
}

type statusInput struct {
	Status string `json:"status"`
}

// SetPlanetSystemModerStatus godoc
// @Summary Установить статус модератора
// @Description Устанавливает статус заявки (Завершена/Отклонена) - только для модераторов и админов
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Param input body statusInput true "Новый статус"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/planet-system/{system_id}/moder [put]
func (h *Handler) SetPlanetSystemModerStatus(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUserRoleVal, okRole := ctx.Get("user_role")
	currentUserIDVal, okID := ctx.Get("user_id")
	if !okRole || !okID {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("требуется авторизация"))
		return
	}
	currentUserRole := currentUserRoleVal.(role.Role)
	if currentUserRole != role.Moderator && currentUserRole != role.Admin {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("недостаточно прав"))
		return
	}

	var status_input statusInput
	if err := ctx.ShouldBindJSON(&status_input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := status_input.Status
	if status != "Завершена" && status != "Отклонена" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("статус может быть изменен только на 'Завершена' или 'Отклонена'"))
		return
	}

	system, err := h.Repository.GetPlanetSystemByID(uint(system_id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if system.Status != "Сформирована" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("только заявки в статусе 'Сформирована' можно завершать/отклонять"))
		return
	}

	var moder_id uint
	switch v := currentUserIDVal.(type) {
	case uint:
		moder_id = v
	case int:
		moder_id = uint(v)
	case float64:
		moder_id = uint(v)
	default:
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("неверный айди модератора: %T", currentUserIDVal))
		return
	}

	if err = h.Repository.SetPlanetSystemModerStatus(uint(system_id), moder_id, status); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Успех")
}

// UpdateSystemResults godoc
// @Summary Обновить результаты расчётов
// @Description Обновляет температуры в Temperature_requests (с псевдо-ключом)
// @Tags planet-system
// @Accept json
// @Produce json
// @Param system_id path int true "ID системы"
// @Param key query string true "Auth key (8 байт)"
// @Param body body []map[string]interface{} true "Results: [{"planet_id": uint, "temperature": uint}]"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/planet-system/{system_id}/results [put]
func (h *Handler) UpdateSystemResults(ctx *gin.Context) {
	systemID, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	key := ctx.Query("key")
	if key != "abc123xy" {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid key"))
		return
	}

	var results []struct {
		PlanetID    uint `json:"planet_id"`
		Temperature uint `json:"temperature"`
	}
	if err := ctx.ShouldBindJSON(&results); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	for _, res := range results {
		if err := h.Repository.UpdateTemperatureRequest(uint(systemID), res.PlanetID, res.Temperature); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
