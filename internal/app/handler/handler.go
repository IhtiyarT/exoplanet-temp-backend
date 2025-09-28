package handler

import (
	"LABS-BMSTU-BACKEND/internal/app/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Minio *minio.Client
}

func NewHandler(r *repository.Repository, m *minio.Client) *Handler {
	return &Handler{
		Repository: r,
		Minio: m,
	}
}

func (h *Handler) RegisterPlanetHandler(router *gin.Engine) {
	router.GET("/", h.GetPlanets)              //список с фильтрацией
	router.GET("/planet/:id", h.GetPlanetById) //одна запись
	router.POST("/planet/", h.CreatePlanet)
	router.PUT("/planet/:id", h.UpdatePlanet)
	router.DELETE("planet/:id", h.DeletePlanet)
	router.POST("/add/:planet_id", h.AddPlanetToSystem)
	router.POST("/image/add/:planet_id", h.AddImage)

}

func (h *Handler) RegisterPlanetSystemHandler(router *gin.Engine) {
	router.POST("/delete", h.DeletePlanetSystem)
}

func (h *Handler) RegisterTemperatureRequestHandler(router *gin.Engine) {
	router.GET("/temps-request/:system_id", h.GetTempRequestData)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources/styles", "./resources/styles")
}

func (h *Handler) successHandler(ctx *gin.Context, key string, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		key:      data,
	})
}

func (h *Handler) successAddHandler(ctx *gin.Context, key string, data interface{}) {
	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		key:      data,
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
