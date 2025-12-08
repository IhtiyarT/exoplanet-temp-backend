package handler

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"LABS-BMSTU-BACKEND/internal/app/redis"
	"LABS-BMSTU-BACKEND/internal/app/role"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type RegisterUserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}

// RegisterUser godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags user
// @Accept json
// @Produce json
// @Param input body RegisterUserReq true "Данные для регистрации"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req RegisterUserReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин и пароль обязательны"))
		return
	}

	candidate, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if candidate.Login == req.Login {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь уже существует"))
		return
	}

	user := ds.Users{
		Login:    req.Login,
		Password: req.Password,
		Role:     role.Role(req.Role),
	}

	if err := h.Repository.CreateUser(user); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Успешно создан",
	})
}

// GetProfile godoc
// @Summary Получить профиль пользователя
// @Description Возвращает информацию о пользователе по ID
// @Tags user
// @Accept json
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/user/{user_id} [get]
func (h *Handler) GetProfile(ctx *gin.Context) {
	currentUserID := ctx.GetUint("user_id")
	currentUserRole := ctx.MustGet("user_role").(role.Role)

	requestedID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный user_id"))
		return
	}

	if currentUserRole == role.User && uint(requestedID) != currentUserID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("недостаточно прав для просмотра чужого профиля"))
		return
	}

	user, err := h.Repository.GetUserByID(uint(requestedID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пользователь не найден"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id": user.UserID,
		"login":   user.Login,
		"role":    user.Role,
	})
}

type UpdateProfileRequest struct {
	NewLogin    string `json:"new_login"`
	NewPassword string `json:"new_password"`
}

// UpdateProfile godoc
// @Summary Обновить профиль пользователя
// @Description Обновляет данные текущего пользователя
// @Tags user
// @Accept json
// @Produce json
// @Param input body UpdateProfileRequest true "Данные для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/user/me [put]
func (h *Handler) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var input UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(userID, input.NewLogin, input.NewPassword); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successHandler(ctx, "message", "данные успешно обновлены")
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя и возвращает JWT токен
// @Tags user
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин и пароль обязательны"))
	}

	user, err := h.Repository.LoginUser(req.Login, req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	token, err := redis.GenerateJWTToken(uint(user.UserID), user.Role)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	err = h.Repository.SaveJWTToken(uint(user.UserID), token)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// Logout godoc
// @Summary Выход из системы
// @Description Выполняет выход пользователя и добавляет токен в черный список
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/user/logout [post]
func (h *Handler) Logout(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("токен не найден"))
		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey(), nil
	})
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный токен"))
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный токен"))
		return
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный токен"))
		return
	}

	expTime := time.Unix(int64(exp), 0)
	now := time.Now()
	durationUntilExp := expTime.Sub(now)

	if durationUntilExp <= 0 {
		h.successHandler(ctx, "message", "успешный выход")
		return
	}

	err = h.Repository.AddToBlacklist(jwtStr, durationUntilExp)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	user_id_float, ok := claims["user_id"].(float64)
	if ok {
		user_id := uint(user_id_float)
		user_id_str := strconv.FormatUint(uint64(user_id), 10)
		h.Repository.DeleteTokenByUserID(user_id_str)
	}

	h.successHandler(ctx, "message", "успешный выход")
}
