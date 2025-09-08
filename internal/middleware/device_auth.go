package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"github.com/labmino/runsight-backend/internal/utils"
)

func DeviceAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := utils.ExtractTokenFromHeader(authHeader)
		
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"message": "Device token required",
				"error_code": "ERR_DEVICE_TOKEN_REQUIRED",
			})
			c.Abort()
			return
		}

		device, err := utils.ValidateDeviceToken(db, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"message": "Invalid device token",
				"error_code": "ERR_INVALID_DEVICE_TOKEN",
				"details": gin.H{"error": err.Error()},
			})
			c.Abort()
			return
		}

		c.Set("device", device)
		c.Set("device_id", device.DeviceID)
		c.Set("user_id", device.UserID)
		c.Next()
	}
}