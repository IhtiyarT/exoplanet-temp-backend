package handler

import (
	"net/http"
	"strconv"
	// "time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) DeletePlanetSystem(ctx *gin.Context) {
	system_id, err := h.Repository.GetDraftPlanetSystemID()
	if system_id == 0 && err == nil{
		logrus.Infof("Система для удаления не найдена")
	}
	h.Repository.DeletePlanetSystem(uint(system_id))
    ctx.Redirect(http.StatusNotFound, "/")
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
	
	ctx.JSON(http.StatusOK, gin.H{
		"planet_system": planet_system,
	})
}
