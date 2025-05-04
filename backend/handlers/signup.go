package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// TODO: Ensure /api/signup is registered before any authentication middleware is applied.
// The actual router setup file (main.go or router.go) needs to be updated to make sure this route is public.

func SignupHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		_, err = db.Exec(
			"INSERT INTO users (email, password_hash) VALUES (?, ?)",
			req.Email, string(hash),
		)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "user created"})
	}
}
