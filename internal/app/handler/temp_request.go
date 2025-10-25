package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeletePlanetFromSystem godoc
// @Summary Удалить планету из системы
// @Description Удаляет планету из планетной системы (только для черновиков и создателя системы)
// @Tags temperature-request
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Param planet_id path int true "ID планеты"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/temperature-req/{system_id}/planet/{planet_id} [delete]
func (h *Handler) DeletePlanetFromSystem(ctx *gin.Context) {
	rawUserID, ok := ctx.Get("user_id")
	if !ok {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("не авторизован"))
		return
	}
	userID := rawUserID.(uint)

	systemID, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid system_id"))
		return
	}

	planetID, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid planet_id"))
		return
	}

	system, err := h.Repository.GetPlanetSystemByID(uint(systemID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система не найдена"))
		return
	}

	if system.UserID != userID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("нельзя изменять чужую систему"))
		return
	}

	if system.Status != "Черновик" {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("можно удалять только из черновика"))
		return
	}

	if err := h.Repository.DeletePlanetFromSystem(uint(systemID), uint(planetID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "планета удалена из системы")
}

type UpdateDistanceInput struct {
	PlanetDistance uint `json:"planet_distance" binding:"required"`
}

// UpdatePlanetDistance godoc
// @Summary Обновить расстояние до планеты
// @Description Обновляет расстояние до планеты в планетной системе (только для черновиков и создателя системы)
// @Tags temperature-request
// @Accept json
// @Produce json
// @Param system_id path int true "ID планетной системы"
// @Param planet_id path int true "ID планеты"
// @Param input body UpdateDistanceInput true "Новое расстояние до планеты"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/temperature-req/{system_id}/planet/{planet_id} [put]
func (h *Handler) UpdatePlanetDistance(ctx *gin.Context) {
	rawUserID, ok := ctx.Get("user_id")
	if !ok {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("не авторизован"))
		return
	}
	userID := rawUserID.(uint)

	systemID, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный system_id"))
		return
	}

	planetID, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный planet_id"))
		return
	}

	system, err := h.Repository.GetPlanetSystemByID(uint(systemID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("система не найдена"))
		return
	}

	if system.UserID != userID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("нельзя изменять чужую систему"))
		return
	}

	if system.Status != "Черновик" {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("можно изменять только черновик"))
		return
	}

	var input UpdateDistanceInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdatePlanetDistance(uint(systemID), uint(planetID), input.PlanetDistance); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "расстояние до планеты успешно обновлено")
}

