package handler

import (
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