package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/apierror"
)

const defaultMaxBodyBytes int64 = 1 << 20 // 1 MiB

// MaxBodyBytes rejects oversized request bodies early.
func MaxBodyBytes(max int64) gin.HandlerFunc {
	if max <= 0 {
		max = defaultMaxBodyBytes
	}
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		}
		c.Next()
		if c.Writer.Status() == http.StatusRequestEntityTooLarge {
			return
		}
		// Gin may surface MaxBytesError during ShouldBindJSON; map if still open.
		if err := c.Errors.Last(); err != nil {
			if _, ok := err.Err.(*http.MaxBytesError); ok {
				ae := apierror.New(http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Request body exceeds size limit.")
				c.AbortWithStatusJSON(ae.HTTPStatus, ae.ToResponse())
			}
		}
	}
}
