package middleware_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/apierror"
	"github.com/quantumshield/qguard/backend/internal/middleware"
)

func TestRecoveryReturnsJSONWithoutStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(log))
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "boom") || strings.Contains(body, "goroutine") {
		t.Fatalf("response leaked internal panic/stack details: %s", body)
	}

	var env apierror.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Error.Code != apierror.CodeInternalError {
		t.Fatalf("code = %s", env.Error.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS([]string{"http://localhost:3000"}))
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestIPRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := middleware.NewIPRateLimiter(1, 1)
	r := gin.New()
	r.Use(lim.Middleware())
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first status = %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", rec2.Code)
	}
}
