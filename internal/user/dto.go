package user

import "github.com/go-playground/validator/v10"

type CreateUserPayload struct {
	FirstName string  `json:"firstName" validate:"required,max=255"`
	LastName  *string `json:"lastName" validate:"omitempty,max=255"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=6,max=20"`
	Role      *string `json:"role" validate:"omitempty,oneof=user host"`
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=20"`
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
