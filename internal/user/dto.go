package user

import "github.com/go-playground/validator/v10"

type CreateUserPayload struct {
	FirstName  string  `json:"firstName" validate:"required,max=255"`
	LastName   *string `json:"lastName" validate:"omitempty,max=255"`
	Email      string  `json:"email" validate:"required,email"`
	Role       *string `json:"role" validate:"omitempty,oneof=user host"`
	IsVerified *bool   `json:"isVerified" validate:"omitempty"`
}

type LoginPayload struct {
	Email string `json:"email" validate:"required,email"`
}

func (payload *CreateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(payload)
}

type ResponseUserDTO struct {
	ID         string  `json:"id" db:"id"`
	FirstName  string  `json:"firstName" db:"first_name"`
	LastName   *string `json:"lastName" db:"last_name"`
	Email      string  `json:"email" db:"email"`
	Role       string  `json:"role" db:"role"`
	IsVerified bool    `json:"isVerified" db:"is_verified"`
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

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
}

type GoogleIDTokenClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Expiry        int64  `json:"exp"`
}
