package main

import (
	"database/sql"
	"egg-tracker/backend/db"
	"egg-tracker/backend/handlers"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func ensureDatabasesWithSampleData() error {
	// Use absolute paths inside the container for persistence
	sqlitePath := "/app/data/eggtracker.db"
	duckdbPath := "/app/data/eggtracker.duckdb"
	needSample := false
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		db, err := db.InitDB(sqlitePath)
		if err != nil {
			return err
		}
		defer db.Close()
		needSample = true
	}
	if needSample {
		db, err := sql.Open("sqlite3", sqlitePath)
		if err != nil {
			return err
		}
		defer db.Close()
		// Insert sample species
		db.Exec(`INSERT OR IGNORE INTO species (name, active) VALUES ('Chicken', 1), ('Goose', 1), ('Guinea Fowl', 1)`)
		// Insert sample egg colors
		db.Exec(`INSERT OR IGNORE INTO egg_colors (name, active) VALUES ('White', 1), ('Brown', 1), ('Blue', 1)`)
		// Insert sample egg sizes
		db.Exec(`INSERT OR IGNORE INTO egg_sizes (name, active) VALUES ('Small', 1), ('Medium', 1), ('Large', 1)`)
		// Insert sample coops
		db.Exec(`INSERT OR IGNORE INTO coops (name, active) VALUES ('Main Coop', 1), ('Back Barn', 1)`)
		// Insert sample eggs
		db.Exec(`INSERT INTO eggs (date_laid, species, deleted, created_at, updated_at) VALUES ('2024-05-01', 'Chicken', 0, '2024-05-01', '2024-05-01'), ('2024-05-02', 'Goose', 0, '2024-05-02', '2024-05-02')`)
		// Insert sample inventory actions
		db.Exec(`INSERT INTO inventory_actions (quantity, species, action, notes, date, created_at, updated_at) VALUES (10, 'Goose', 'collected', 'note', '2024-05-01', '2024-05-01', '2024-05-01')`)
	}
	// Always ensure DuckDB exists and is up to date
	if _, err := os.Stat(duckdbPath); os.IsNotExist(err) || needSample {
		// Run full ETL to create DuckDB from SQLite
		if err := handlers.EtlFullRefreshFromMainWithPaths(sqlitePath, duckdbPath); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := ensureDatabasesWithSampleData(); err != nil {
		log.Fatalf("failed to initialize databases: %v", err)
	}

	database, err := db.InitDB("/app/data/eggtracker.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize the database schema and seed data
	if err := db.InitializeDatabase("/app/data/eggtracker.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := gin.Default()

	// Public routes
	router.POST("/api/signup", handlers.SignupHandler(database))

	// --- CORS middleware (must be first) ---
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		AllowOrigins: []string{
			"http://frontend:3000",
			"http://backend:8080",
			"http://localhost:3000",
			"http://localhost:8080",
			// "http://192.168.1.42:3000", // add your ip if you're testing on a different machine on the local network
			// "http://192.168.1.42:8080",
		},
	}))

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/api/login", handlers.LoginHandler(database))
	router.POST("/api/refresh", handlers.RefreshHandler())

	inv := router.Group("/api/inventory")
	{
		inv.POST("", handlers.CreateInventoryHandler(database))
		inv.GET("", handlers.ListInventoryHandler(database))
		inv.PUT("/:id", handlers.UpdateInventoryHandler(database))
		inv.DELETE("/:id", handlers.DeleteInventoryHandler(database))
	}

	// Register /api/options endpoints
	options := router.Group("/api/options")
	{
		options.GET("/:type", handlers.ListOptionsHandler(database))
		options.POST("/:type", handlers.AddOptionHandler(database))
		options.PUT("/:type/:id", handlers.EditOptionHandler(database))
		options.POST("/:type/:id/deactivate", handlers.DeactivateOptionHandler(database))
		options.POST("/:type/:id/reactivate", handlers.ReactivateOptionHandler(database))
	}

	// Register /api/reports endpoint
	router.GET("/api/reports", handlers.ReportsHandler())

	// Register ETL full refresh endpoint
	router.POST("/api/etl/full", handlers.FullETLHandler(database))

	// Register backup endpoint
	router.POST("/api/backup", handlers.BackupHandler())

	router.Run("0.0.0.0:8080")
}
