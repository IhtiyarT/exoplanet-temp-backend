package handler

import (
	"fmt"
	"net/http"
	"strconv"

	// "strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeletePlanetFromSystem(ctx *gin.Context) {
    system_id, err := strconv.Atoi(ctx.Param("system_id"))
    if err != nil {
        h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid system_id"))
        return
    }

    planet_id, err := strconv.Atoi(ctx.Param("planet_id"))
    if err != nil {
        h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid planet_id"))
        return
    }

    if err := h.Repository.DeletePlanetFromSystem(uint(system_id), uint(planet_id)); err != nil {
        h.errorHandler(ctx, http.StatusInternalServerError, err)
        return
    }

    h.successHandler(ctx, "message", "планета удалено из системы")
}

type UpdateDistanceInput struct {
	PlanetDistance uint `json:"planet_distance" binding:"required"`
}

func (h *Handler) UpdatePlanetDistance(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный system_id"))
		return
	}

	planet_id, err := strconv.Atoi(ctx.Param("planet_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный planet_id"))
		return
	}

	var input UpdateDistanceInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdatePlanetDistance(uint(system_id), uint(planet_id), input.PlanetDistance); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "расстояние до планеты успешно обновлено")
}

