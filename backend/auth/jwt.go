package auth

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	AccessSecret  = []byte("your-access-secret")  // Replace with env/config in production
	RefreshSecret = []byte("your-refresh-secret") // Replace with env/config in production
)

func GenerateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessSecret)
}

func GenerateRefreshToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(RefreshSecret)
}

func ParseRefreshToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id")
	}
	return int64(userID), nil
}

func ParseAccessToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return AccessSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id")
	}
	return int64(userID), nil
}

// Gin middleware to protect routes
// Usage: router.Use(auth.AuthMiddleware())
func AuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		userID, err := ParseAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
