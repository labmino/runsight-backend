package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"go.uber.org/zap"

	"github.com/labmino/runsight-backend/internal/utils"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

func (rl *RateLimiter) CleanupOldLimiters() {
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for key, limiter := range rl.limiters {
				if limiter.Tokens() == float64(rl.burst) {
					delete(rl.limiters, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

func RateLimitMiddleware(requestsPerSecond int, burstSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), burstSize)
	limiter.CleanupOldLimiters()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := limiter.GetLimiter(clientIP)

		if !limiter.Allow() {
			requestID := c.GetString("RequestID")
			
			utils.Warn("Rate limit exceeded",
				zap.String("request_id", requestID),
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded", gin.H{
				"error_code": "ERR_RATE_LIMIT",
				"retry_after": "60", // seconds
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func StrictRateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	requestsPerSecond := float64(requestsPerMinute) / 60.0
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), 2)
	limiter.CleanupOldLimiters()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := limiter.GetLimiter(clientIP)

		if !limiter.Allow() {
			requestID := c.GetString("RequestID")
			
			utils.Warn("Strict rate limit exceeded",
				zap.String("request_id", requestID),
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded for sensitive endpoint", gin.H{
				"error_code": "ERR_STRICT_RATE_LIMIT",
				"retry_after": "120",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}