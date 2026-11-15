// Package handler maps HTTP requests to application services.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/apierror"
	"github.com/quantumshield/qguard/backend/internal/service"
)

// API groups HTTP handlers for /api/v1.
type API struct {
	meta *service.MetaService
}

// NewAPI constructs the API handler group.
func NewAPI(meta *service.MetaService) *API {
	return &API{meta: meta}
}

// Health handles GET /api/v1/health.
func (a *API) Health(c *gin.Context) {
	c.JSON(http.StatusOK, a.meta.Health())
}

// Ready handles GET /api/v1/ready.
func (a *API) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, a.meta.Ready())
}

// Info handles GET /api/v1/info.
func (a *API) Info(c *gin.Context) {
	c.JSON(http.StatusOK, a.meta.Info())
}

// NotFound is the fallback JSON 404 handler.
func NotFound(c *gin.Context) {
	err := apierror.NotFound("The requested resource was not found.")
	c.JSON(err.HTTPStatus, err.ToResponse())
}

// MethodNotAllowed is the fallback JSON 405 handler.
func MethodNotAllowed(c *gin.Context) {
	err := apierror.New(http.StatusMethodNotAllowed, apierror.CodeInvalidRequest, "Method not allowed for this endpoint.")
	c.JSON(err.HTTPStatus, err.ToResponse())
}

// AbortWithAPIError writes a structured API error.
func AbortWithAPIError(c *gin.Context, err *apierror.Error) {
	c.AbortWithStatusJSON(err.HTTPStatus, err.ToResponse())
}
