package handler

import (
	"LABS-BMSTU-BACKEND/internal/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) DeletePlanetSystem(ctx *gin.Context) {
	system_id, err := h.Repository.GetDraftPlanetSystemID()
	if system_id == 0 && err == nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("система для удаления не найдена"))
		return
	}
	h.Repository.DeletePlanetSystem(uint(system_id))
	h.successHandler(ctx, "message", "успешно удалено")
}

func (h *Handler) GetPlanetSystemByID(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	planet_system, err := h.Repository.GetPlanetSystemByID(uint(id))
	if err != nil {
		logrus.Error(err)
	}

	h.successHandler(ctx, "planet_system", planet_system)
}

func (h *Handler) GetPlanetSystemDraftID(ctx *gin.Context) {
	system_id, err := h.Repository.GetDraftPlanetSystemID()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
	}

	planet_count, err := h.Repository.GetCountBySystemID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
	}

	planet_system, err := h.Repository.GetPlanetSystemAndPlanetsByID(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
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

func (h *Handler) GetPlanetSystemsList(ctx *gin.Context) {
	if system_id := ctx.Query("system_id"); system_id != "" {
		system_id_int, err := strconv.Atoi(system_id)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		h.Repository.GetPlanetSystemByID(uint(system_id_int))
		return
	}

	system_status := ctx.Query("system_status")
	start_date_str := ctx.Query("start_date")
	end_date_str := ctx.Query("end_date")

	if start_date_str == "" {
		start_date_str = "0001-01-01"
	}
	if end_date_str == "" {
		end_date_str = time.Now().Format("2006-01-02")
	}

	start_date, err_start := utils.ParseDateString(start_date_str)
	end_date, err_end := utils.ParseDateString(end_date_str)

	if err_end != nil || err_start != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("incorrect `start_date` or `end_date`"))
		return
	}

	planet_systems, err := h.Repository.GetPlanetSystemsList(system_status, start_date, end_date)
	if err != nil {
		h.errorHandler(ctx, http.StatusNoContent, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"systems_count": len(planet_systems),
		"planet_systems": planet_systems,
	})
}

func (h *Handler) GetPlanetSystemAndPlanetsByID(ctx *gin.Context) {
	system_id := ctx.Param("system_id")
	id, err := strconv.Atoi(system_id)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet_system, err := h.Repository.GetPlanetSystemAndPlanetsByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
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
		"planet_count": planet_count,
	})
}

type PlanetSystemInput struct {
	StarType       string `json:"star_type"`
	StarName       string `json:"star_name"`
	StarLuminosity uint   `json:"star_luminosity"`
}

func (h *Handler) UpdatePlanetSystem(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
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

func (h *Handler) SetPlanetSystemFormed(ctx *gin.Context) {
	system_id, err := strconv.ParseUint(ctx.Param("system_id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	id := uint(system_id)

	if err := h.Repository.SetPlanetSystemFormed(id, h.getUserID()); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "успех",
	})
}

type statusInput struct {
	Status string `json:"status"`
}

func (h *Handler) SetPlanetSystemModerStatus(ctx *gin.Context) {
	system_id, err := strconv.Atoi(ctx.Param("system_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var status_input statusInput
	if err := ctx.ShouldBindJSON(&status_input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := status_input.Status

	if status != "Завершена" && status != "Отклонена" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("статус может быть изменен только на завершена или отклонена"))
		return
	}

	if err = h.Repository.SetPlanetSystemModerStatus(uint(system_id), h.getModerID(), status); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Успех")
}
