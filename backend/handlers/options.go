package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"egg-tracker/backend/models"

	"github.com/gin-gonic/gin"
)

type OptionInput struct {
	Name string `json:"name" binding:"required"`
}

func getOptionTable(optionType string) (string, bool) {
	switch strings.ToLower(optionType) {
	case "species":
		return "species", true
	case "eggcolor":
		return "egg_colors", true
	case "eggsize":
		return "egg_sizes", true
	case "coop":
		return "coops", true
	default:
		return "", false
	}
}

func ListOptionsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeStr := c.Param("type")
		table, ok := getOptionTable(typeStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option type"})
			return
		}
		rows, err := db.Query("SELECT id, name, active, created_at, updated_at FROM " + table + " ORDER BY name ASC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()
		var options []models.OptionBase
		for rows.Next() {
			var opt models.OptionBase
			if err := rows.Scan(&opt.ID, &opt.Name, &opt.Active, &opt.CreatedAt, &opt.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
				return
			}
			options = append(options, opt)
		}
		c.JSON(http.StatusOK, options)
	}
}

func AddOptionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeStr := c.Param("type")
		table, ok := getOptionTable(typeStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option type"})
			return
		}
		var input OptionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		res, err := db.Exec("INSERT INTO "+table+" (name, active) VALUES (?, 1)", input.Name)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "option already exists"})
			return
		}
		id, _ := res.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func EditOptionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeStr := c.Param("type")
		table, ok := getOptionTable(typeStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option type"})
			return
		}
		id := c.Param("id")
		var input OptionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		_, err := db.Exec("UPDATE "+table+" SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", input.Name, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	}
}

func DeactivateOptionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeStr := c.Param("type")
		table, ok := getOptionTable(typeStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option type"})
			return
		}
		id := c.Param("id")
		_, err := db.Exec("UPDATE "+table+" SET active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deactivated"})
	}
}

func ReactivateOptionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeStr := c.Param("type")
		table, ok := getOptionTable(typeStr)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option type"})
			return
		}
		id := c.Param("id")
		_, err := db.Exec("UPDATE "+table+" SET active = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "reactivated"})
	}
}
