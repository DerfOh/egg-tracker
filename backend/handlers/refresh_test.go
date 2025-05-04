package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"egg-tracker/backend/auth"

	"github.com/gin-gonic/gin"
)

func TestRefreshHandler_IssuesAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/refresh", RefreshHandler())

	// Generate a valid refresh token for user ID 42
	refreshToken, err := auth.GenerateRefreshToken(42)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	req, _ := http.NewRequest("POST", "/api/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "access_token") {
		t.Fatalf("expected access_token in response, got: %s", w.Body.String())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr))))
}
