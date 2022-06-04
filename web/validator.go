package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var DefaultValidator *validator.Validate = NewJSONValidator()

func NewJSONValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			return field.Name
		}

		name := strings.SplitN(jsonTag, ",", 2)[0]

		if name == "-" {
			name = ""
		}

		return name
	})

	return v
}

func ValidateStruct(s interface{}) (bool, []Cause) {
	return ValidateStructWith(DefaultValidator, s)
}

func ValidateVar(name string, value interface{}, tag string) (bool, []Cause) {
	return ValidateVarWith(DefaultValidator, name, value, tag)
}

func ValidateStructWith(v *validator.Validate, s interface{}) (bool, []Cause) {
	if err := v.Struct(s); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var causes []Cause

		for _, fieldError := range validationErrors {
			fieldName := extractFieldName(fieldError)
			causes = append(causes, FieldError(fieldName, fieldError.Tag()))
		}

		return false, causes
	}

	return true, nil
}

func ValidateVarWith(v *validator.Validate, name string, value interface{}, tag string) (bool, []Cause) {
	if err := v.Var(value, tag); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var causes []Cause

		for _, fieldError := range validationErrors {
			causes = append(causes, FieldError(name, fieldError.Tag()))
		}

		return false, causes
	}

	return true, nil
}

func ValidationError(c *gin.Context, causes ...Cause) {
	ErrorResponse(c, http.StatusBadRequest, "Validation Error", causes...)
}

func extractFieldName(fieldError validator.FieldError) string {
	return strings.Join(strings.Split(fieldError.Namespace(), ".")[1:], ".")
}
