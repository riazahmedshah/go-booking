package errs

import "net/http"

const (
	CodePropertyHeld            = "PROPERTY_HELD"
	CodeBookingInProgress       = "BOOKING_IN_PROGRESS"
	CodePaymentFailed           = "PAYMENT_FAILED"
	CodeBookingAlreadyConfirmed = "BOOKING_ALREADY_CONFIRMED"
)

var (
	ErrPropertyHeld = &AppError{
		StatusCode: http.StatusConflict,
		Code:       CodePropertyHeld,
		Message:    "this property is currently held by another user, please try again shortly",
		Op:         "createBooking.lockCheck",
	}

	ErrBookingInProgress = &AppError{
		StatusCode: http.StatusConflict,
		Code:       CodeBookingInProgress,
		Message:    "you already have a booking request in progress for this property",
		Op:         "createBooking.lockCheck",
	}

	ErrBookingAlreadyConfirmed = &AppError{
		StatusCode: http.StatusConflict,
		Code:       CodeBookingAlreadyConfirmed,
		Message:    "this booking has already been confirmed",
		Op:         "confirmBooking.idempotencyCheck",
	}

	ErrPropertyUnavailable = &AppError{
		StatusCode: http.StatusConflict,
		Code:       "PROPERTY_UNAVAILABLE",
		Message:    "property is unavailable for the selected dates",
		Op:         "createBooking.dateCheck",
	}

	// ErrBookingNotFound = New(
	// 	http.StatusNotFound,
	// 	"booking record not found",
	// 	nil,
	// )
)
