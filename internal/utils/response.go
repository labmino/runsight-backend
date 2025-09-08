package utils

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	requestID := c.GetString("RequestID")
	
	Info("API Success Response",
		zap.String("request_id", requestID),
		zap.String("message", message),
		zap.Int("status_code", statusCode),
	)
	
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	requestID := c.GetString("RequestID")
	
	Error("API Error Response",
		zap.String("request_id", requestID),
		zap.String("message", message),
		zap.Int("status_code", statusCode),
		zap.Any("error", err),
	)
	
	c.JSON(statusCode, Response{
		Status:  "error",
		Message: message,
		Error:   err,
	})
}

func ValidationErrorResponse(c *gin.Context, err error) {
	var validationErrors []string
	
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErr {
			validationErrors = append(validationErrors, fieldErr.Field()+" is "+fieldErr.Tag())
		}
	} else {
		validationErrors = append(validationErrors, err.Error())
	}
	
	requestID := c.GetString("RequestID")
	
	Warn("Validation Error",
		zap.String("request_id", requestID),
		zap.Strings("validation_errors", validationErrors),
		zap.String("original_error", err.Error()),
	)
	
	c.JSON(http.StatusBadRequest, Response{
		Status:  "error",
		Message: "Validation failed",
		Error:   validationErrors,
	})
}

func GenerateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}