package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
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
	
	c.JSON(http.StatusBadRequest, Response{
		Status:  "error",
		Message: "Validation failed",
		Error:   validationErrors,
	})
}