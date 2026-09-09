package errs

import "fmt"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Op         string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s]: %s %v", e.Op, e.Message, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Op, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(statusCode int, internalErr error, code, message, op string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Op:         op,
		Err:        internalErr,
	}
}
