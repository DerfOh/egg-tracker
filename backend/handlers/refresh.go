package handlers

import (
	"net/http"

	"egg-tracker/backend/auth"

	"github.com/gin-gonic/gin"
)

func RefreshHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil || refreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
			return
		}

		userID, err := auth.ParseRefreshToken(refreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		accessToken, err := auth.GenerateAccessToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
	}
}
