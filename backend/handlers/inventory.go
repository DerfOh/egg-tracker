package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"egg-tracker/backend/models"

	"github.com/gin-gonic/gin"
)

type InventoryInput struct {
	Quantity int     `json:"quantity" binding:"required"`
	Species  string  `json:"species" binding:"required"`
	Coop     string  `json:"coop" binding:"required"`
	EggColor string  `json:"egg_color" binding:"required"`
	EggSize  string  `json:"egg_size" binding:"required"`
	Action   string  `json:"action" binding:"required"`
	Notes    *string `json:"notes"`
	Date     string  `json:"date" binding:"required"` // ISO8601 date
}

func CreateInventoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input InventoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		date, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
			return
		}
		res, err := db.Exec(
			"INSERT INTO inventory_actions (quantity, species, coop, egg_color, egg_size, action, notes, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			input.Quantity, input.Species, input.Coop, input.EggColor, input.EggSize, input.Action, input.Notes, date,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func ListInventoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, quantity, species, coop, egg_color, egg_size, action, notes, date, created_at, updated_at FROM inventory_actions")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()
		var actions []models.InventoryAction
		for rows.Next() {
			var act models.InventoryAction
			var notes sql.NullString
			var coop, eggColor, eggSize sql.NullString
			if err := rows.Scan(&act.ID, &act.Quantity, &act.Species, &coop, &eggColor, &eggSize, &act.Action, &notes, &act.Date, &act.CreatedAt, &act.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
				return
			}
			if notes.Valid {
				act.Notes = &notes.String
			}
			if coop.Valid {
				act.Coop = coop.String
			}
			if eggColor.Valid {
				act.EggColor = eggColor.String
			}
			if eggSize.Valid {
				act.EggSize = eggSize.String
			}
			actions = append(actions, act)
		}
		c.JSON(http.StatusOK, actions)
	}
}

func UpdateInventoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input InventoryInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		date, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
			return
		}
		_, err = db.Exec(
			"UPDATE inventory_actions SET quantity = ?, species = ?, coop = ?, egg_color = ?, egg_size = ?, action = ?, notes = ?, date = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			input.Quantity, input.Species, input.Coop, input.EggColor, input.EggSize, input.Action, input.Notes, date, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	}
}

func DeleteInventoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM inventory_actions WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
