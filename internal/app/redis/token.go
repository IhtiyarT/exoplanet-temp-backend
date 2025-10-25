package redis

import (
	"LABS-BMSTU-BACKEND/internal/app/role"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func jwtKey() []byte {
	k := os.Getenv("JWT_KEY")
	if k == "" {
		k = "secret-key"
	}
	return []byte(k)
}

func GenerateJWTToken(user_id uint, role role.Role) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user_id
	claims["role"] = int(role)
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString(jwtKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}