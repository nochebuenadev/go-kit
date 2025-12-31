/*
Package check provides a simple and consistent interface for struct validation.

It wraps the go-playground/validator package and maps its results to the application's
standard apperr.AppErr type, providing consistent error reporting and human-readable
messages in Spanish.

Example usage:

	type User struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"min=18"`
	}

	user := User{Email: "invalid-email", Age: 16}
	err := check.Global().Struct(user)
	if err != nil {
		fmt.Println(err.Error()) // Output: El campo 'Email' debe ser un correo electrónico válido
	}
*/
package check
