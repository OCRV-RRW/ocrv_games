package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
)

var validate = validator.New()

type ApiError struct {
	Errors []*ErrorResponse `json:"errors"`
}

type ErrorResponse struct {
	Parameter string `json:"parameter"`
	Message   string `json:"message"`
}

func GetErrorResponse(parameter string, tag string) *ErrorResponse {
	errorResponse := &ErrorResponse{}
	errorResponse.Parameter = parameter
	if tag == "required" {
		errorResponse.Message = fmt.Sprintf("The %s field is required.", parameter)
	} else if tag == "email" {
		errorResponse.Message = fmt.Sprintf("The %s field has an incorrect email address.", parameter)
	} else {
		errorResponse.Message = fmt.Sprintf("The %s field is incorrect.", parameter)
	}

	return errorResponse
}

func ValidateStruct[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldName := GetJSONTag(payload, err.Field())
			if fieldName == "" {
				continue
			}
			errors = append(errors, GetErrorResponse(fieldName, err.Tag()))
		}
	}
	return errors
}

func GetJSONTag(structure interface{}, fieldName string) string {
	val := reflect.ValueOf(structure)
	typ := val.Type()

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	field, found := typ.FieldByName(fieldName)
	if !found {
		return ""
	}

	jsonTag := field.Tag.Get("json")
	return jsonTag
}
