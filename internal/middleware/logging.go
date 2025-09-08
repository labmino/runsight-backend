package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/labmino/runsight-backend/internal/utils"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get or set request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetString("RequestID")
		}
		if requestID == "" {
			requestID = "unknown"
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.Int("body_size", bodySize),
			zap.String("user_agent", userAgent),
		}

		// Add user information if available
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		if userEmail, exists := c.Get("user_email"); exists {
			fields = append(fields, zap.String("user_email", userEmail.(string)))
		}

		// Add device information if available
		if _, exists := c.Get("device"); exists {
			fields = append(fields, zap.String("device_id", "masked"))
		}

		// Add error information if available
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		if statusCode >= 500 {
			utils.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			utils.Warn("HTTP Request", fields...)
		} else {
			utils.Info("HTTP Request", fields...)
		}
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateRequestID()
		}
		
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}