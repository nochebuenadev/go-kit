package check

import (
	"strings"
	"testing"

	"github.com/nochebuenadev/go-kit/pkg/apperr"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18,max=120"`
}

func TestGlobal(t *testing.T) {
	v := Global()
	if v == nil {
		t.Fatal("Global() returned nil")
	}

	v2 := Global()
	if v != v2 {
		t.Error("Global() should return the same singleton instance")
	}
}

func TestValidator_Struct(t *testing.T) {
	v := Global()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errCode string
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Email: "john@example.com",
				Age:   30,
			},
			wantErr: true,
			errCode: "INVALID_ARGUMENT",
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			wantErr: true,
			errCode: "INVALID_ARGUMENT",
		},
		{
			name: "below min age",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   17,
			},
			wantErr: true,
			errCode: "INVALID_ARGUMENT",
		},
		{
			name: "above max age",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   121,
			},
			wantErr: true,
			errCode: "INVALID_ARGUMENT",
		},
		{
			name:    "invalid input type (not a struct)",
			input:   "not a struct",
			wantErr: true,
			errCode: "INTERNAL_ERROR", // validator.InvalidValidationError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Struct() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				ae, ok := err.(*apperr.AppErr)
				if !ok {
					t.Fatalf("expected *apperr.AppErr, got %T", err)
				}
				if ae.GetCode() != tt.errCode {
					t.Errorf("expected error code %s, got %s", tt.errCode, ae.GetCode())
				}

				// Verify Spanish message for some cases
				if tt.name == "missing required field" && !strings.HasPrefix(ae.Error(), "INVALID_ARGUMENT: El campo 'Name' es obligatorio") {
					t.Errorf("unexpected error message: %s", ae.Error())
				}
			}
		})
	}
}
