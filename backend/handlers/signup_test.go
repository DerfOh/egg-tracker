package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"egg-tracker/backend/db"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/api/signup", SignupHandler(db))
	return r
}

func TestSignupHandler_CreatesUser(t *testing.T) {
	testDBPath := "test_signup.db"
	defer os.Remove(testDBPath)

	database, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	router := setupTestRouter(database)

	payload := SignupRequest{
		Email:    "test@example.com",
		Password: "supersecret",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
	}

	// Check user exists in DB
	var count int
	row := database.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", payload.Email)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}
}
