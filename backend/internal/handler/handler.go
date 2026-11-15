// Package handler maps HTTP requests to application services.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/apierror"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/etherscan"
	"github.com/quantumshield/qguard/backend/internal/repository"
	"github.com/quantumshield/qguard/backend/internal/service"
)

// API groups HTTP handlers for /api/v1.
type API struct {
	meta    *service.MetaService
	account *service.AccountService
	scan    *service.ScanService
}

// NewAPI constructs the API handler group.
func NewAPI(meta *service.MetaService, account *service.AccountService, scan *service.ScanService) *API {
	return &API{meta: meta, account: account, scan: scan}
}

// Health handles GET /api/v1/health.
func (a *API) Health(c *gin.Context) {
	c.JSON(http.StatusOK, a.meta.Health())
}

// Ready handles GET /api/v1/ready.
func (a *API) Ready(c *gin.Context) {
	resp := a.meta.Ready(c.Request.Context())
	status := http.StatusOK
	if !resp.Ready {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, resp)
}

// Info handles GET /api/v1/info.
func (a *API) Info(c *gin.Context) {
	c.JSON(http.StatusOK, a.meta.Info())
}

// GetAccount handles GET /api/v1/accounts/:address and /api/v1/wallet/:address.
func (a *API) GetAccount(c *gin.Context) {
	if a.account == nil {
		AbortWithAPIError(c, apierror.Internal("Account service unavailable."))
		return
	}
	view, err := a.account.GetAccount(c.Request.Context(), c.Param("address"))
	if err != nil {
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusOK, view)
}

// GetAccountTransactions handles GET /api/v1/accounts/:address/transactions.
func (a *API) GetAccountTransactions(c *gin.Context) {
	if a.account == nil {
		AbortWithAPIError(c, apierror.Internal("Account service unavailable."))
		return
	}
	hist, err := a.account.GetHistory(c.Request.Context(), c.Param("address"))
	if err != nil {
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	raw, err := service.MarshalTxHistoryJSON(hist)
	if err != nil {
		AbortWithAPIError(c, apierror.Internal("Failed to encode history response."))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}

// GetAccountExposure handles GET .../exposure (accounts or wallet).
func (a *API) GetAccountExposure(c *gin.Context) {
	addr := c.Param("address")
	// Prefer latest completed scan exposure when available.
	if a.scan != nil {
		if exp, err := a.scan.LatestExposure(c.Request.Context(), addr); err == nil {
			c.JSON(http.StatusOK, exp)
			return
		} else if !errors.Is(err, repository.ErrNotFound) {
			var inv blockchain.ErrInvalidAddress
			if errors.As(err, &inv) {
				AbortWithAPIError(c, apierror.BadRequest(apierror.CodeInvalidRequest, inv.Error()))
				return
			}
		}
	}
	if a.account == nil {
		AbortWithAPIError(c, apierror.Internal("Account service unavailable."))
		return
	}
	report, err := a.account.GetExposure(c.Request.Context(), addr)
	if err != nil {
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetWalletRisk handles GET /api/v1/wallet/:address/risk.
func (a *API) GetWalletRisk(c *gin.Context) {
	if a.scan == nil {
		AbortWithAPIError(c, apierror.Internal("Scan service unavailable."))
		return
	}
	risk, err := a.scan.LatestRisk(c.Request.Context(), c.Param("address"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			AbortWithAPIError(c, apierror.NotFound("No completed risk assessment found for this address. Start a scan first."))
			return
		}
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusOK, risk)
}

type startScanRequest struct {
	Address string `json:"address"`
	ChainID *int64 `json:"chain_id"`
}

// StartScan handles POST /api/v1/scan.
func (a *API) StartScan(c *gin.Context) {
	if a.scan == nil {
		AbortWithAPIError(c, apierror.Internal("Scan service unavailable."))
		return
	}
	var req startScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		AbortWithAPIError(c, apierror.BadRequest(apierror.CodeInvalidRequest, "Request body must be JSON with an address field."))
		return
	}
	if req.Address == "" {
		AbortWithAPIError(c, apierror.BadRequest(apierror.CodeInvalidRequest, "address is required"))
		return
	}
	job, err := a.scan.StartScan(c.Request.Context(), req.Address)
	if err != nil {
		if errors.Is(err, repository.ErrCapacity) {
			AbortWithAPIError(c, apierror.New(http.StatusServiceUnavailable, "CAPACITY_EXCEEDED", "In-memory scan capacity exceeded; retry later or restart the process."))
			return
		}
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusAccepted, job)
}

// GetScan handles GET /api/v1/scan/:id.
func (a *API) GetScan(c *gin.Context) {
	if a.scan == nil {
		AbortWithAPIError(c, apierror.Internal("Scan service unavailable."))
		return
	}
	job, err := a.scan.GetScan(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			AbortWithAPIError(c, apierror.NotFound("Scan not found."))
			return
		}
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusOK, job)
}

// GetReport handles GET /api/v1/reports/:id.
func (a *API) GetReport(c *gin.Context) {
	if a.scan == nil {
		AbortWithAPIError(c, apierror.Internal("Scan service unavailable."))
		return
	}
	rep, err := a.scan.GetReport(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			AbortWithAPIError(c, apierror.NotFound("Report not found."))
			return
		}
		AbortWithAPIError(c, mapDomainError(err))
		return
	}
	c.JSON(http.StatusOK, rep)
}

func mapDomainError(err error) *apierror.Error {
	var inv blockchain.ErrInvalidAddress
	if errors.As(err, &inv) {
		return apierror.BadRequest(apierror.CodeInvalidRequest, inv.Error())
	}
	var mismatch blockchain.ErrChainMismatch
	if errors.As(err, &mismatch) {
		return apierror.BadRequest(apierror.CodeUnsupportedChain, mismatch.Error())
	}
	var ae *etherscan.APIError
	if errors.As(err, &ae) {
		switch ae.Code {
		case etherscan.CodeInvalidAddr:
			return apierror.BadRequest(apierror.CodeInvalidRequest, ae.Message)
		case etherscan.CodeRateLimited:
			return apierror.New(http.StatusTooManyRequests, "RATE_LIMITED", ae.Message)
		default:
			return apierror.New(http.StatusBadGateway, "HISTORY_PROVIDER_ERROR", ae.Message)
		}
	}
	return apierror.New(http.StatusBadGateway, "RPC_UNAVAILABLE", err.Error())
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
