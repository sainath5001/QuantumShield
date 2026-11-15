// Package middleware provides Gin middleware for the Q-Guard API.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quantumshield/qguard/backend/internal/apierror"
)

const (
	// RequestIDHeader is the HTTP header used for request correlation.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the Gin context key for the request ID.
	RequestIDKey = "request_id"
)

// RequestID assigns or propagates a request ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// Logger emits structured request logs.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		reqID, _ := c.Get(RequestIDKey)
		log.Info("request",
			"request_id", reqID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// Recovery catches panics and returns a JSON error without stack traces to clients.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID, _ := c.Get(RequestIDKey)
				log.Error("panic recovered",
					"request_id", reqID,
					"error", rec,
				)
				err := apierror.Internal("An unexpected error occurred.")
				c.AbortWithStatusJSON(err.HTTPStatus, err.ToResponse())
			}
		}()
		c.Next()
	}
}

// CORS applies a simple allow-list CORS policy.
func CORS(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Expose-Headers", RequestIDHeader)
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Timeout attaches a per-request context deadline. Handlers should respect c.Request.Context().
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() == context.DeadlineExceeded && c.Writer.Status() == 0 {
			err := apierror.Timeout("Request timed out.")
			c.AbortWithStatusJSON(err.HTTPStatus, err.ToResponse())
		}
	}
}
