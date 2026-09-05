package user

import "github.com/go-playground/validator/v10"

type CreateUserPayload struct {
	FirstName string  `json:"firstName" validate:"required,max=255"`
	LastName  *string `json:"lastName" validate:"omitempty,max=255"`
	Email     string  `json:"email" validate:"required,email"`
	Role      *string `json:"role" validate:"omitempty,oneof=user host"`
}

type LoginPayload struct {
	Email string `json:"email" validate:"required,email"`
}

func (payload *CreateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(payload)
}

type ResponseUserDTO struct {
	ID        string  `json:"id" db:"id"`
	FirstName string  `json:"firstName" db:"first_name"`
	LastName  *string `json:"lastName" db:"last_name"`
	Email     string  `json:"email" db:"email"`
	Role      string  `json:"role" db:"role"`
}

type SendOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
	OTP   int64  `json:"otp" validate:"required"`
}

type VerifyOTPResult struct {
	UserExists bool   `json:"userExists"`
	Email      string `json:"email"`
}

type SessionData struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
