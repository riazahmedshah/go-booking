package errs

import "net/http"

var (
	ErrDuplicateEmail = &AppError{
		StatusCode: http.StatusConflict,
		Code:       "DUPLICATE_EMAIL",
		Message:    "email already alreadt taken, please use different email",
		Op:         "createUser.emailCheck",
	}

	ErrUserNotFound = &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "USER_NOT_FOUND",
		Message:    "user not found",
		Op:         "getUser.idCheck",
	}

	//
	ErrInvalidOTP = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_OTP",
		Message:    "invalid OTP provided",
		Op:         "verifyOTP.otpCheck",
	}

	ErrInvalidOTPFormat = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "MALFORMED_OTP",
		Message:    "malformed/invalid OTP",
		Op:         "verifyOTP.otpFormatCheck",
	}

	ErrEmailNotVerified = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "EMAIL_NOT_VERIFIED",
		Message:    "email not verified, please verify email first",
		Op:         "register.emailVerificationCheck",
	}
)
