package etherscan

import (
	"context"
	"sync"
	"time"
)

// RateLimiter gates outbound provider calls.
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// TokenBucket is a simple token-bucket rate limiter.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	last       time.Time
	now        func() time.Time
}

// NewTokenBucket creates a limiter with the given sustained RPS and burst size.
func NewTokenBucket(rps float64, burst int) *TokenBucket {
	if rps <= 0 {
		rps = 4
	}
	if burst < 1 {
		burst = 1
	}
	return &TokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rps,
		last:       time.Now(),
		now:        time.Now,
	}
}

// Wait blocks until a token is available or the context is canceled.
func (t *TokenBucket) Wait(ctx context.Context) error {
	for {
		t.mu.Lock()
		now := t.now()
		elapsed := now.Sub(t.last).Seconds()
		if elapsed > 0 {
			t.tokens += elapsed * t.refillRate
			if t.tokens > t.maxTokens {
				t.tokens = t.maxTokens
			}
			t.last = now
		}
		if t.tokens >= 1 {
			t.tokens--
			t.mu.Unlock()
			return nil
		}
		need := 1 - t.tokens
		wait := time.Duration(need / t.refillRate * float64(time.Second))
		t.mu.Unlock()
		if wait < time.Millisecond {
			wait = time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// NopLimiter never blocks (tests).
type NopLimiter struct{}

func (NopLimiter) Wait(context.Context) error { return nil }
