package middleware

import (
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/labmino/runsight-backend/internal/utils"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		
		c.Header("X-Content-Type-Options", "nosniff")
		
		c.Header("X-XSS-Protection", "1; mode=block")
		
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
		
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	}
}

func InputSanitization() gin.HandlerFunc {
	var (
		sqlInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
		xssPattern         = regexp.MustCompile(`(?i)(<script|<iframe|<object|<embed|<link|<meta|javascript:|vbscript:|onload|onerror|onclick|onmouseover)`)
		pathTraversalPattern = regexp.MustCompile(`(\.\./|\.\.\|/\.\./|\.\.\\)`)
	)

	return func(c *gin.Context) {
		requestID := c.GetString("RequestID")
		
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsMaliciousInput(value, sqlInjectionPattern, xssPattern, pathTraversalPattern) {
					utils.Warn("Malicious input detected in query parameter",
						zap.String("request_id", requestID),
						zap.String("parameter", key),
						zap.String("value", value),
						zap.String("client_ip", c.ClientIP()),
					)
					
					utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input detected", gin.H{
						"error_code": "ERR_INVALID_INPUT",
						"field": key,
					})
					c.Abort()
					return
				}
			}
		}

		path := c.Request.URL.Path
		if containsMaliciousInput(path, pathTraversalPattern) {
			utils.Warn("Malicious input detected in path",
				zap.String("request_id", requestID),
				zap.String("path", path),
				zap.String("client_ip", c.ClientIP()),
			)
			
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid path", gin.H{
				"error_code": "ERR_INVALID_PATH",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func containsMaliciousInput(input string, patterns ...*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func SanitizeString(input string) string {
	sanitized := html.EscapeString(input)
	
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, contentType := range allowedTypes {
		allowedMap[contentType] = true
	}

	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			
			if idx := strings.Index(contentType, ";"); idx != -1 {
				contentType = contentType[:idx]
			}
			contentType = strings.TrimSpace(contentType)

			if !allowedMap[contentType] {
				requestID := c.GetString("RequestID")
				
				utils.Warn("Invalid content type",
					zap.String("request_id", requestID),
					zap.String("content_type", contentType),
					zap.String("client_ip", c.ClientIP()),
				)
				
				utils.ErrorResponse(c, http.StatusUnsupportedMediaType, "Unsupported content type", gin.H{
					"error_code": "ERR_UNSUPPORTED_MEDIA_TYPE",
					"allowed_types": allowedTypes,
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

func MaxRequestSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			requestID := c.GetString("RequestID")
			
			utils.Warn("Request too large",
				zap.String("request_id", requestID),
				zap.Int64("content_length", c.Request.ContentLength),
				zap.Int64("max_size", maxSize),
				zap.String("client_ip", c.ClientIP()),
			)
			
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "Request too large", gin.H{
				"error_code": "ERR_REQUEST_TOO_LARGE",
				"max_size_bytes": maxSize,
			})
			c.Abort()
			return
		}
		
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		
		c.Next()
	}
}