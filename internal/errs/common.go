package errs

import "net/http"

const (
	CodeValidationError = "VALIDATION_ERROR"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeDuplicate       = "DUPLICATE"
)

func Internal(message, op string, internalErr error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeInternalError,
		Message:    message,
		Op:         op,
		Err:        internalErr,
	}
}
