package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/models"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}

	return New(config.Config{
		Port:         "8080",
		DatabasePath: ":memory:",
		JWTSecret:    "test-secret",
		FrontendURL:  "http://localhost:5173",
	}, db, nil)
}

func doJSONRequest(router *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func getTokenFromLogin(t *testing.T, router *gin.Engine) string {
	t.Helper()
	w := doJSONRequest(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "student001",
		"password": "password123",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login failed, status=%d body=%s", w.Code, w.Body.String())
	}

	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	return body.Data.Token
}

func TestRegisterAndLogin(t *testing.T) {
	r := setupTestRouter(t)

	registerResp := doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "student001",
		"password": "password123",
		"email":    "student001@example.com",
	}, "")
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d: %s", registerResp.Code, registerResp.Body.String())
	}

	loginResp := doJSONRequest(r, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "student001",
		"password": "password123",
	}, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", loginResp.Code, loginResp.Body.String())
	}

	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	if body.Data.Token == "" {
		t.Fatal("expected non-empty jwt token")
	}
}

func TestJWTProtectedRoute(t *testing.T) {
	r := setupTestRouter(t)

	unauthorizedResp := doJSONRequest(r, http.MethodGet, "/api/v1/users/me", nil, "")
	if unauthorizedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", unauthorizedResp.Code)
	}

	doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "student001",
		"password": "password123",
	}, "")
	token := getTokenFromLogin(t, r)

	authorizedResp := doJSONRequest(r, http.MethodGet, "/api/v1/users/me", nil, token)
	if authorizedResp.Code != http.StatusOK {
		t.Fatalf("expected authorized status 200, got %d: %s", authorizedResp.Code, authorizedResp.Body.String())
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	r := setupTestRouter(t)
	doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "student001",
		"password": "password123",
	}, "")
	token := getTokenFromLogin(t, r)

	adminResp := doJSONRequest(r, http.MethodGet, "/api/v1/admin/dashboard", nil, token)
	if adminResp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status 403, got %d: %s", adminResp.Code, adminResp.Body.String())
	}
}
