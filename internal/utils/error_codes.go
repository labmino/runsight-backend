package utils

// Standard error codes for the API
const (
	// Authentication & Authorization
	ErrInvalidCredentials = "ERR_INVALID_CREDENTIALS"
	ErrTokenExpired      = "ERR_TOKEN_EXPIRED"
	ErrTokenInvalid      = "ERR_TOKEN_INVALID"
	ErrUnauthorized      = "ERR_UNAUTHORIZED"
	ErrForbidden         = "ERR_FORBIDDEN"
	
	// Validation & Input
	ErrValidationFailed     = "ERR_VALIDATION_FAILED"
	ErrInvalidInput        = "ERR_INVALID_INPUT"
	ErrInvalidPath         = "ERR_INVALID_PATH"
	ErrMissingField        = "ERR_MISSING_FIELD"
	ErrInvalidFormat       = "ERR_INVALID_FORMAT"
	
	// Resources
	ErrResourceNotFound    = "ERR_RESOURCE_NOT_FOUND"
	ErrResourceExists      = "ERR_RESOURCE_EXISTS"
	ErrResourceConflict    = "ERR_RESOURCE_CONFLICT"
	
	// Database
	ErrDatabaseConnection  = "ERR_DATABASE_CONNECTION"
	ErrDatabaseQuery       = "ERR_DATABASE_QUERY"
	ErrDatabaseTransaction = "ERR_DATABASE_TRANSACTION"
	
	// Device & Pairing
	ErrDeviceNotFound      = "ERR_DEVICE_NOT_FOUND"
	ErrDeviceAlreadyPaired = "ERR_DEVICE_ALREADY_PAIRED"
	ErrPairingCodeInvalid  = "ERR_PAIRING_CODE_INVALID"
	ErrPairingCodeExpired  = "ERR_PAIRING_CODE_EXPIRED"
	
	// Rate Limiting
	ErrRateLimit           = "ERR_RATE_LIMIT"
	ErrStrictRateLimit     = "ERR_STRICT_RATE_LIMIT"
	
	// Request Processing
	ErrRequestTimeout      = "ERR_REQUEST_TIMEOUT"
	ErrRequestTooLarge     = "ERR_REQUEST_TOO_LARGE"
	ErrUnsupportedMediaType = "ERR_UNSUPPORTED_MEDIA_TYPE"
	
	// Server Errors
	ErrInternal           = "ERR_INTERNAL"
	ErrServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"
	ErrCircuitBreakerOpen = "ERR_CIRCUIT_BREAKER_OPEN"
)

// DetailedErrorResponse represents a structured error response with error codes
type DetailedErrorResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	ErrorCode string                 `json:"error_code"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// GetErrorMessage returns a human-readable message for error codes
func GetErrorMessage(errorCode string) string {
	messages := map[string]string{
		ErrInvalidCredentials:   "Invalid email or password",
		ErrTokenExpired:        "Authentication token has expired",
		ErrTokenInvalid:        "Invalid authentication token",
		ErrUnauthorized:        "Authentication required",
		ErrForbidden:          "Access denied",
		
		ErrValidationFailed:    "Request validation failed",
		ErrInvalidInput:       "Invalid input provided",
		ErrInvalidPath:        "Invalid request path",
		ErrMissingField:       "Required field is missing",
		ErrInvalidFormat:      "Invalid data format",
		
		ErrResourceNotFound:   "Requested resource not found",
		ErrResourceExists:     "Resource already exists",
		ErrResourceConflict:   "Resource conflict",
		
		ErrDatabaseConnection: "Database connection failed",
		ErrDatabaseQuery:      "Database query failed",
		ErrDatabaseTransaction: "Database transaction failed",
		
		ErrDeviceNotFound:     "Device not found",
		ErrDeviceAlreadyPaired: "Device is already paired",
		ErrPairingCodeInvalid: "Invalid pairing code",
		ErrPairingCodeExpired: "Pairing code has expired",
		
		ErrRateLimit:          "Rate limit exceeded",
		ErrStrictRateLimit:    "Rate limit exceeded for sensitive operation",
		
		ErrRequestTimeout:     "Request timeout",
		ErrRequestTooLarge:    "Request payload too large",
		ErrUnsupportedMediaType: "Unsupported media type",
		
		ErrInternal:           "Internal server error",
		ErrServiceUnavailable: "Service temporarily unavailable",
		ErrCircuitBreakerOpen: "Service temporarily unavailable due to high error rate",
	}
	
	if message, exists := messages[errorCode]; exists {
		return message
	}
	return "An error occurred"
}