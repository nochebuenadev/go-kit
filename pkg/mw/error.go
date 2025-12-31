package mw

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// AppErrorHandler returns an Echo HTTPErrorHandler that standardizes error responses.
// It maps apperr.AppErr to matching HTTP status codes and logs the errors using the provided logger.
func AppErrorHandler(logger logz.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var appErr *apperr.AppErr
		if !errors.As(err, &appErr) {
			var he *echo.HTTPError
			if errors.As(err, &he) {
				appErr = apperr.Internal("%s", he.Error()).WithError(err)
			} else {
				appErr = apperr.Internal("error inesperado").WithError(err)
			}
		}

		logger.With(
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
			"method", mGetRequestMethod(c),
			"uri", mGetRequestURI(c),
			"code", appErr.GetCode(),
		).LogError("PETICIÓN_FALLIDA", appErr)

		status := mapAppErrToHTTPStatus(appErr)

		_ = c.JSON(status, appErr)
	}
}

// mapAppErrToHTTPStatus maps an application error code to a standard HTTP status code.
func mapAppErrToHTTPStatus(appErr *apperr.AppErr) int {
	switch apperr.ErrorCode(appErr.GetCode()) {
	case apperr.ErrInvalidInput:
		return http.StatusBadRequest
	case apperr.ErrResourceNotFound:
		return http.StatusNotFound
	case apperr.ErrConflict:
		return http.StatusConflict
	case apperr.ErrUnauthorized:
		return http.StatusUnauthorized
	case apperr.ErrPermissionDenied:
		return http.StatusForbidden
	case apperr.ErrInternal:
		return http.StatusInternalServerError
	case apperr.ErrNotImplemented:
		return http.StatusNotImplemented
	case apperr.ErrUnavailable:
		return http.StatusServiceUnavailable
	case apperr.ErrDeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// mGetRequestMethod safely retrieves the request method from echo.Context.
func mGetRequestMethod(c echo.Context) string {
	if c.Request() == nil {
		return "UNKNOWN"
	}
	return c.Request().Method
}

// mGetRequestURI safely retrieves the request URI from echo.Context.
func mGetRequestURI(c echo.Context) string {
	if c.Request() == nil {
		return "UNKNOWN"
	}
	return c.Request().RequestURI
}
