package validation

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/riazahmedshah/stayz/internal/errs"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// This magic snippet makes the validator look at `json:"fieldName"` tags!
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{Validator: v}
}

func (cv *CustomValidator) Validate(i any) error {
	err := cv.Validator.Struct(i)
	if err == nil {
		return nil
	}

	castedErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errs.Internal("unexpected validation error", "validator.Validate", err)
	}

	var msgs []string
	for _, fieldErr := range castedErrors {
		fieldName := fieldErr.Field()

		var message string
		switch fieldErr.Tag() {
		case "required":
			message = "is required"
		case "email":
			message = "must be a valid email"
		case "max":
			message = "exceeds max length of " + fieldErr.Param()
		case "oneof":
			message = "must be one of: " + fieldErr.Param()
		default:
			message = "failed " + fieldErr.Tag() + " validation"
		}

		msgs = append(msgs, fieldName+" "+message)
	}

	return &errs.AppError{
		StatusCode: http.StatusBadRequest,
		Code:       errs.CodeValidationError,
		Message:    strings.Join(msgs, "; "),
		Op:         "validator.Validate",
	}
}
