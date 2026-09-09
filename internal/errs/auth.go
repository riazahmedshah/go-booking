package errs

import "net/http"

const (
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
)

var (
	ErrUnauthorized = &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       CodeUnauthorized,
		Message:    "please log in to continue",
		Op:         "auth.requireAuth",
	}

	ErrForbidden = &AppError{
		StatusCode: http.StatusForbidden,
		Code:       CodeForbidden,
		Message:    "you do not have permission to access this resource",
		Op:         "auth.requireRole",
	}
)
