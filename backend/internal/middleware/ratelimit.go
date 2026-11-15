package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/apierror"
)

// IPRateLimiter is a simple per-IP token-bucket gate for inbound HTTP requests.
// Intended for local MVP DoS softening — not a substitute for edge rate limiting.
type IPRateLimiter struct {
	mu            sync.Mutex
	visitors      map[string]*visitor
	rps           float64
	burst         int
	cleanupOnce   sync.Once
}

type visitor struct {
	tokens     float64
	last       time.Time
	lastSeen   time.Time
}

// NewIPRateLimiter constructs a limiter. rps<=0 disables limiting (passthrough).
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	if burst < 1 {
		burst = 1
	}
	return &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rps,
		burst:    burst,
	}
}

// Middleware returns a Gin handler that rate-limits by client IP.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	if l == nil || l.rps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	l.cleanupOnce.Do(func() { go l.cleanupLoop() })
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.allow(ip) {
			err := apierror.New(http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests from this client. Retry later.")
			c.AbortWithStatusJSON(err.HTTPStatus, err.ToResponse())
			return
		}
		c.Next()
	}
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	v, ok := l.visitors[ip]
	if !ok {
		l.visitors[ip] = &visitor{tokens: float64(l.burst - 1), last: now, lastSeen: now}
		return true
	}
	elapsed := now.Sub(v.last).Seconds()
	v.tokens += elapsed * l.rps
	if v.tokens > float64(l.burst) {
		v.tokens = float64(l.burst)
	}
	v.last = now
	v.lastSeen = now
	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, v := range l.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}
