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
	"golang.org/x/crypto/bcrypt"
)

func setupLoginTestDB() (*sql.DB, func()) {
	testDBPath := "test_login.db"
	dbase, _ := sql.Open("sqlite3", testDBPath)
	db.Migrate(dbase)
	// Insert a user
	pw, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	dbase.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", "login@example.com", string(pw))
	return dbase, func() {
		dbase.Close()
		os.Remove(testDBPath)
	}
}

func TestLoginHandler_IssuesTokens(t *testing.T) {
	dbase, cleanup := setupLoginTestDB()
	defer cleanup()

	router := gin.Default()
	router.POST("/api/login", LoginHandler(dbase))

	payload := LoginRequest{
		Email:    "login@example.com",
		Password: "testpass123",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == "" {
		t.Fatalf("expected access_token in response")
	}
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected refresh_token cookie to be set and HTTPOnly")
	}
}
