package handler

import (
	"fmt"
	"net/http"
	"strconv"

	// "strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	// "LABS-BMSTU-BACKEND/internal/app/ds"
)

func (h *Handler) GetTempRequestData(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		logrus.Error(err)
	}

	if !h.Repository.CheckIsDelete(uint(system_id)) {
		logrus.Infof("Страница уже удалена")
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	temp_datas, err := h.Repository.GetPlanetsWithSystemData(uint(system_id))
	if err != nil {
		logrus.Error(err)
	}
	
	cart_count, err := h.Repository.GetCountBySystemID(uint(system_id))
    if err != nil {
        cart_count = 0
        logrus.Error("Error getting cart count:", err)
    }

	ctx.JSON(http.StatusOK, gin.H{
		"temp_datas": temp_datas,
		"cart_count": cart_count,
	})
}

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

