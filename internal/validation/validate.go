package validation

import (
	"Games/internal/api"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
)

var validate = validator.New()

func GetErrorResponse(parameter string, tag string) *api.Error {
	errorResponse := &api.Error{}
	errorResponse.Parameter = parameter
	if tag == "required" {
		errorResponse.Message = fmt.Sprintf("The %s field is required.", parameter)
	} else if tag == "email" {
		errorResponse.Message = fmt.Sprintf("The %s field has an incorrect email address.", parameter)
	} else {
		errorResponse.Message = fmt.Sprintf("The %s field is incorrect.", parameter)
	}
	errorResponse.Code = api.IncorrectParameter

	return errorResponse
}

func ValidateStruct[T any](payload T) []*api.Error {
	var errors []*api.Error
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
