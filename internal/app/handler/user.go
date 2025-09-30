package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	IsAdmin  *bool  `json:"is_admin"` // указатель, чтобы отличать "не передано" от false (Возможно стоит переделать)
}

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var input map[string]interface{}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if input["login"] == nil || input["password"] == nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("login и password обязательны"))
		return
	}

	if err := h.Repository.CreateUser(input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "Успешно зарегестрирован")
}

func (h *Handler) GetProfile(ctx *gin.Context) {
	user_id_str, err := ctx.Cookie("user_id")
	if err != nil || user_id_str == "" {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("не авторизован"))
		return
	}

	user_id, _ := strconv.Atoi(user_id_str)
	user, err := h.Repository.GetUserByID(uint(user_id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id":  user.UserID,
	})
}

func (h *Handler) UpdateProfile(ctx *gin.Context) {
	user_id_str, err := ctx.Cookie("user_id")
	if err != nil || user_id_str == "" {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("не авторизован"))
		return
	}

	user_id, _ := strconv.Atoi(user_id_str)

	var input map[string]string
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(uint(user_id), input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "данные успешно обновлены")
}

func (h *Handler) Login(ctx *gin.Context) {
	var input map[string]string
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.AuthenticateUser(input["login"], input["password"])
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
		return
	}

	// user_id будет лежать в куки
	ctx.SetCookie("user_id", fmt.Sprintf("%d", user.UserID), 3600, "/", "", false, true)

	h.successHandler(ctx, "message", "успешная авторизация")
}

func (h *Handler) Logout(ctx *gin.Context) {
	ctx.SetCookie("user_id", "", -1, "/", "", false, true)
	h.successHandler(ctx, "message", "успешный выход")
}
