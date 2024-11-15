package validation

import (
	"Games/internal/validation/error_code"
	"fmt"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ApiError struct {
	Errors []*ErrorResponse `json:"errors"`
}

type ErrorResponse struct {
	Parameter string `json:"parameter"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

func ValidateStruct[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Parameter = err.StructNamespace()
			element.Code = error_code.INVALID_PARAMETER
			if err.Tag() == "required" {
				element.Message = fmt.Sprintf("The %s field is required.", err.StructNamespace())
			} else if err.Tag() == "email" {
				element.Message = fmt.Sprintf("The %s field has an incorrect email address.", err.StructNamespace())
			} else {
				element.Message = fmt.Sprintf("The %s field is incorrect.", err.StructNamespace())
			}
			errors = append(errors, &element)
		}
	}
	return errors
}
