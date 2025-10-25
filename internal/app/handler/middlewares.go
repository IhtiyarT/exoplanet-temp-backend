package handler

import (
	"LABS-BMSTU-BACKEND/internal/app/role"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

func jwtKey() []byte {
	k := os.Getenv("JWT_KEY")
	if k == "" {
		k = "secret-key"
	}
	return []byte(k)
}

func (h *Handler) WithAuthCheck(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "требуется авторизация"})
			return
		}

		jwtStr = strings.TrimPrefix(jwtStr, jwtPrefix)

		inBlacklist, err := h.Repository.IsInBlacklist(jwtStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "ошибка проверки токена"})
			return
		}
		if inBlacklist {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "токен недействителен"})
			return
		}

		token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret-key"), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "некорректный токен"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "ошибка токена"})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "некорректный user_id"})
			return
		}

		userRole, ok := claims["role"].(float64)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "некорректная роль"})
			return
		}

		ctx.Set("user_id", uint(userID))
		ctx.Set("user_role", role.Role(userRole))

		if len(allowedRoles) > 0 {
			authorized := false
			for _, allowed := range allowedRoles {
				if allowed == role.Role(userRole) {
					authorized = true
					break
				}
			}
			if !authorized {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "недостаточно прав"})
				return
			}
		}

		ctx.Next()
	}
}

func (h *Handler) WithOptionalAuthCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")
		if jwtStr == "" || !strings.HasPrefix(jwtStr, jwtPrefix) {
			ctx.Next()
			return
		}

		jwtStr = strings.TrimPrefix(jwtStr, jwtPrefix)

		inBlacklist, err := h.Repository.IsInBlacklist(jwtStr)
		if err != nil {
			ctx.Next()
			return
		}
		if inBlacklist {
			ctx.Next()
			return
		}

		token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret-key"), nil
		})
		if err != nil || !token.Valid {
			ctx.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.Next()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			ctx.Next()
			return
		}

		userRole, ok := claims["role"].(float64)
		if !ok {
			ctx.Next()
			return
		}

		ctx.Set("user_id", uint(userID))
		ctx.Set("user_role", role.Role(userRole))

		ctx.Next()
	}
}
