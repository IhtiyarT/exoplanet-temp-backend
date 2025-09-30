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
	Minio      *minio.Client
}

func NewHandler(r *repository.Repository, m *minio.Client) *Handler {
	return &Handler{
		Repository: r,
		Minio:      m,
	}
}

func (h *Handler) RegisterPlanetHandler(router *gin.Engine) {
	router.GET("api/planet", h.GetPlanets)
	router.GET("api/planet/:id", h.GetPlanetById)
	router.POST("api/planet/", h.CreatePlanet)
	router.PUT("api/planet/:id", h.UpdatePlanet)
	router.DELETE("api/planet/:id", h.DeletePlanet)
	router.POST("api/add/:planet_id", h.AddPlanetToSystem)
	router.POST("api/image/add/:planet_id", h.AddImage)
}

func (h *Handler) RegisterPlanetSystemHandler(router *gin.Engine) {
	router.GET("api/planet-system/draft/id", h.GetPlanetSystemDraftID)
	router.GET("api/planet-system/list", h.GetPlanetSystemsList)
	router.GET("api/planet-system/:system_id", h.GetPlanetSystemAndPlanetsByID)
	router.PUT("api/planet-system/:system_id", h.UpdatePlanetSystem)
	router.PUT("api/planet-system/:system_id/form", h.SetPlanetSystemFormed)
	router.PUT("api/planet-system/:system_id/moder", h.SetPlanetSystemModerStatus)
	router.POST("api/planet-system/delete", h.DeletePlanetSystem)
}

func (h *Handler) RegisterTemperatureRequestHandler(router *gin.Engine) {
	router.DELETE("api/temperature-req/:system_id/planet/:planet_id", h.DeletePlanetFromSystem)
	router.PUT("api/temperature-req/:system_id/planet/:planet_id", h.UpdatePlanetDistance)
}

func (h *Handler) RegisterUserHandler(router *gin.Engine) {
	router.POST("api/user/register", h.RegisterUser)
	router.GET("api/user/me", h.GetProfile)
	router.PUT("api/user/me", h.UpdateProfile)
	router.POST("api/user/login", h.Login)
	router.POST("api/user/logout", h.Logout)
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

func (h *Handler) getUserID() uint {
	return 1
}

func (h *Handler) getModerID() uint {
	return 2
}
