/*
Package logz provides a flexible and structured logging utility for Go applications.

It wraps the standard library's slog package to provide a consistent interface (Logger)
that supports different log levels, structured data, context-aware logging, and
integration with custom error types like apperr.AppErr.

Example usage:

	logz.MustInit()
	logger := logz.Global()

	logger.Info("logz: iniciando aplicación", "version", "1.0.0")

	if err != nil {
		logger.LogError("logz: fallo al procesar petición", err)
	}
*/
package logz
