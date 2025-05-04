package handlers

import (
	"database/sql"
	"net/http"

	"egg-tracker/backend/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		var id int64
		var hash string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", req.Email).Scan(&id, &hash)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		accessToken, err := auth.GenerateAccessToken(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
			return
		}
		refreshToken, err := auth.GenerateRefreshToken(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
			return
		}

		c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true) // HTTPOnly
		c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
	}
}
