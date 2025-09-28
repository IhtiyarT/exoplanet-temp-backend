package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"LABS-BMSTU-BACKEND/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetPlanets(ctx *gin.Context) {
	var planets []ds.Planets
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		planets, err = h.Repository.GetPlanets()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		planets, err = h.Repository.GetPlanetsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	system_id, err1 := h.Repository.GetDraftPlanetSystemID()
	if err1 != nil {
		logrus.Error(err1)
	}

	cartCount, err := h.Repository.GetCountBySystemID(system_id)
	if err != nil {
		cartCount = 0
		logrus.Error("Error getting cart count:", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"planets":   planets,
		"query":     searchQuery,
		"cartCount": cartCount,
		"system_id": system_id,
	})
}

func (h *Handler) GetPlanetById(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	planet, err := h.Repository.GetPlanet(id)
	if err != nil {
		logrus.Error(err)
	}

	h.successHandler(ctx, "planet", planet)
}

func (h *Handler) AddPlanetToSystem(ctx *gin.Context) {
	planet_id, err1 := strconv.Atoi(ctx.Param("planet_id"))
	system_id, err2 := h.Repository.GetDraftPlanetSystemID()
	if err1 != nil {
		logrus.Error(err1)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	if err2 != nil {
		logrus.Error(err2)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	if system_id == 0 {
		system_id, err2 = h.Repository.CreateNewDraftPlanetSystem(uint(1))
		if err2 != nil {
			logrus.Error(err2)
		}
	}

	h.Repository.AddPlanetToSystem(uint(planet_id), uint(system_id))
	ctx.Redirect(http.StatusFound, "/")
}

type PlanetInput struct {
	PlanetTitle string  `json:"planet_title"`
	Description string  `json:"description"`
	Albedo      float64 `json:"albedo"`
	IsDelete    bool    `json:"is_delete"`
}

func (h *Handler) CreatePlanet(ctx *gin.Context) {
	var new_planet PlanetInput
	if err := ctx.ShouldBindJSON(&new_planet); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	planet := ds.Planets{
		PlanetTitle:       new_planet.PlanetTitle,
		PlanetDescription: new_planet.Description,
		Albedo:            new_planet.Albedo,
		IsDelete:          false,
	}

	planet_id, err := h.Repository.CreatePlanet(planet)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "planet_id", planet_id)
}

func (h *Handler) UpdatePlanet(ctx *gin.Context) {
	planet_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var planet_input PlanetInput
	if err := ctx.ShouldBindJSON(&planet_input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	updated_planet, err := h.Repository.UpdatePlanet(uint(planet_id), planet_input)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "planet", updated_planet)
}

func (h *Handler) DeletePlanet(ctx *gin.Context) {
	planet_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeletePlanet(uint(planet_id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Deleted")
}

func (h *Handler) AddImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	planet_id := ctx.Param("planet_id")

	if planet_id == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("planet id not found"))
		return
	}
	if header == nil || header.Size == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("header id not found"))
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	defer func(file multipart.File) {
		errLol := file.Close()
		if errLol != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, errLol)
			return
		}
	}(file)

	// Upload the image to minio server.
	newImageURL, errorCode, errImage := h.createPlanetImage(&file, header, planet_id)
	if errImage != nil {
		h.errorHandler(ctx, errorCode, errImage)
		return
	}

	h.successAddHandler(ctx, "planet_image", newImageURL)
}

// Функция записи фото в минио
func (h *Handler) createPlanetImage(
	file *multipart.File,
	header *multipart.FileHeader,
	planet_id string,
) (string, int, error) {
	newImageURL, errMinio := h.createImageInMinio(file, header)
	if errMinio != nil {
		return "", http.StatusInternalServerError, errMinio
	}
	if err := h.Repository.UpdatePlanetImage(planet_id, newImageURL); err != nil {
		return "", http.StatusInternalServerError, err
	}
	return newImageURL, 0, nil
}
