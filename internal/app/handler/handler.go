package handler

import (
	"LABS-BMSTU-BACKEND/internal/app/repository"
	"LABS-BMSTU-BACKEND/internal/app/role"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.RegisterPlanetHandler(router)
	h.RegisterPlanetSystemHandler(router)
	h.RegisterTemperatureRequestHandler(router)
	h.RegisterUserHandler(router)
}

func (h *Handler) RegisterPlanetHandler(router *gin.Engine) {
	router.GET("api/planet", h.WithOptionalAuthCheck(), h.GetPlanets)
	router.GET("api/planet/:planet_id", h.WithOptionalAuthCheck(), h.GetPlanetById)
	router.POST("api/planet", h.WithAuthCheck(role.Moderator, role.Admin), h.CreatePlanet)
	router.PUT("api/planet/:planet_id", h.WithAuthCheck(role.Moderator, role.Admin), h.UpdatePlanet)
	router.DELETE("api/planet/:planet_id", h.WithAuthCheck(role.Moderator, role.Admin), h.DeletePlanet)
	router.POST("api/planet/add/:planet_id", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.AddPlanetToSystem)
	router.POST("api/planet/image/add/:planet_id", h.WithAuthCheck(role.Moderator, role.Admin), h.AddImage)
}

func (h *Handler) RegisterPlanetSystemHandler(router *gin.Engine) {
    router.GET("api/planet-system/draft/id", h.WithAuthCheck(role.User), h.GetPlanetSystemDraftID)

    router.GET("api/planet-system/list", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.GetPlanetSystemsList)
    router.GET("api/planet-system/:system_id", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.GetPlanetSystemAndPlanetsByID)
    router.PUT("api/planet-system/:system_id", h.WithAuthCheck(role.User), h.UpdatePlanetSystem)
    
    router.PUT("api/planet-system/:system_id/form", h.WithAuthCheck(role.User), h.SetPlanetSystemFormed)
    router.PUT("api/planet-system/:system_id/moder", h.WithAuthCheck(role.Moderator, role.Admin), h.SetPlanetSystemModerStatus)
	router.PUT("/api/planet-system/:system_id/results", h.UpdateSystemResults)
    
    router.DELETE("api/planet-system/delete", h.WithAuthCheck(role.User), h.DeletePlanetSystem)
}

func (h *Handler) RegisterTemperatureRequestHandler(router *gin.Engine) {
	router.DELETE("api/temperature-req/:system_id/planet/:planet_id", h.WithAuthCheck(role.User), h.DeletePlanetFromSystem)
	router.PUT("api/temperature-req/:system_id/planet/:planet_id", h.WithAuthCheck(role.User), h.UpdatePlanetDistance)
}


func (h *Handler) RegisterUserHandler(router *gin.Engine) {
	router.POST("api/user/register", h.RegisterUser)
	router.GET("api/user/:user_id", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.GetProfile)
	router.PUT("api/user/me", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.UpdateProfile)
	router.POST("api/user/login", h.Login)
	router.POST("api/user/logout", h.WithAuthCheck(role.User, role.Moderator, role.Admin), h.Logout)
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
