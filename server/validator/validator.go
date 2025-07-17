package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func Init() {
	v = validator.New()
}

func Validate(obj interface{}) error {
	err := v.Struct(obj)
	if err == nil {
		return nil
	}

	errs := err.(validator.ValidationErrors)
	errorMessages := make([]string, 0, len(errs))
	structName := getType(obj)

	for _, err := range errs {
		fieldName := strings.ReplaceAll(err.Namespace(), structName+".", "")
		tagName := err.Tag()
		switch tagName {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required", getFieldName(fieldName)))
		case "min":
			minLength := err.Param()
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters long", getFieldName(fieldName), minLength))
		case "max":
			maxLength := err.Param()
			errorMessages = append(errorMessages, fmt.Sprintf("%s cannot be more than %s characters long", getFieldName(fieldName), maxLength))
		case "email":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is not a valid email address", getFieldName(fieldName)))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("%s is invalid", getFieldName(fieldName)))
		}
	}
	return errors.New(strings.Join(errorMessages, "; "))
}

func ValidateVariable(obj interface{}, tags, parameterName string) error {
	err := v.Var(obj, tags)
	if err == nil {
		return nil
	}

	return err
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func getFieldName(fieldName string) string {
	fieldName = strings.ReplaceAll(fieldName, "_", " ")
	return strings.Title(fieldName)
}
