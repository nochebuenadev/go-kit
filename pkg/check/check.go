package check

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
)

type (
	// Validator defines the interface for struct validation.
	Validator interface {
		// Struct validates a struct and returns an apperr.AppErr if validation fails.
		Struct(i interface{}) error
	}

	// customValidator is the concrete implementation of Validator using go-playground/validator.
	customValidator struct {
		v *validator.Validate
	}
)

var (
	global Validator
	once   sync.Once
)

// MustInit initializes the global validator instance once.
func MustInit() {
	once.Do(func() {
		v := validator.New()

		global = &customValidator{v: v}
	})
}

// Global returns the singleton validator instance. It initializes it if it's not already.
func Global() Validator {
	if global == nil {
		MustInit()
	}
	return global
}

// Struct implements the Validator interface.
// It maps go-playground validation errors to our standard apperr.AppErr format.
func (cv *customValidator) Struct(i interface{}) error {
	err := cv.v.Struct(i)
	if err == nil {
		return nil
	}

	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return apperr.Wrap(apperr.ErrInternal, "Error interno de validación", err)
	}

	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		firstErr := validateErrs[0]
		msg := getErrorMessage(firstErr)

		return apperr.New(apperr.ErrInvalidInput, msg).
			WithContext("field", firstErr.Field()).
			WithContext("tag", firstErr.Tag()).
			WithError(err)
	}

	return apperr.Wrap(apperr.ErrInvalidInput, "Datos de entrada no válidos", err)
}

// getErrorMessage maps validator.FieldError tags to human-readable error messages in Spanish.
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("El campo '%s' es obligatorio", err.Field())
	case "email":
		return fmt.Sprintf("El campo '%s' debe ser un correo electrónico válido", err.Field())
	case "min":
		return fmt.Sprintf("El campo '%s' es demasiado corto (mínimo %s)", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("El campo '%s' es demasiado largo (máximo %s)", err.Field(), err.Param())
	default:
		return fmt.Sprintf("Error en el campo '%s': regla '%s' no cumplida", err.Field(), err.Tag())
	}
}
