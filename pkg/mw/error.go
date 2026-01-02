package mw

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nochebuenadev/go-kit/pkg/apperr"
	"github.com/nochebuenadev/go-kit/pkg/logz"
)

// AppErrorHandler returns an Echo HTTPErrorHandler that standardizes error responses.
// It maps apperr.AppErr to matching HTTP status codes and logs the errors using the provided logger.
// It also handles echo.HTTPError by mapping them to apperr.AppErr.
func AppErrorHandler(logger logz.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var appErr *apperr.AppErr
		if !errors.As(err, &appErr) {
			var he *echo.HTTPError
			if errors.As(err, &he) {
				appErr = mapHTTPStatusToAppErr(he)
			} else {
				appErr = apperr.Internal("error inesperado").WithError(err)
			}
		}

		if appErr.GetCode() == string(apperr.ErrResourceNotFound) {
			logger.With(
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", mGetRequestMethod(c),
				"uri", mGetRequestURI(c),
				"code", appErr.GetCode(),
			).Warn("mw: recurso no encontrado")
		} else {
			logger.With(
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", mGetRequestMethod(c),
				"uri", mGetRequestURI(c),
				"code", appErr.GetCode(),
			).LogError("mw: petición fallida", appErr)
		}

		status := mapAppErrToHTTPStatus(appErr)

		_ = c.JSON(status, appErr)
	}
}

// mapHTTPStatusToAppErr converts an echo.HTTPError into our standardized apperr.AppErr.
// It maps various HTTP status codes to the most appropriate internal ErrorCode.
func mapHTTPStatusToAppErr(he *echo.HTTPError) *apperr.AppErr {
	var code apperr.ErrorCode

	switch he.Code {
	case http.StatusNotFound:
		code = apperr.ErrResourceNotFound
	case http.StatusUnauthorized:
		code = apperr.ErrUnauthorized
	case http.StatusForbidden:
		code = apperr.ErrPermissionDenied
	case http.StatusBadRequest,
		http.StatusUnprocessableEntity,
		http.StatusRequestEntityTooLarge:
		code = apperr.ErrInvalidInput
	case http.StatusConflict:
		code = apperr.ErrConflict
	case http.StatusMethodNotAllowed:
		code = apperr.ErrNotImplemented
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		code = apperr.ErrDeadlineExceeded
	case http.StatusServiceUnavailable:
		code = apperr.ErrUnavailable
	case http.StatusNotImplemented:
		code = apperr.ErrNotImplemented
	case http.StatusInternalServerError:
		code = apperr.ErrInternal
	default:
		if he.Code >= 500 {
			code = apperr.ErrInternal
		} else {
			code = apperr.ErrInvalidInput
		}
	}

	return apperr.New(code, fmt.Sprintf("%v", he.Message)).WithError(he)
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
// It returns "UNKNOWN" if the request is not present.
func mGetRequestMethod(c echo.Context) string {
	if c.Request() == nil {
		return "UNKNOWN"
	}
	return c.Request().Method
}

// mGetRequestURI safely retrieves the request URI from echo.Context.
// It returns "UNKNOWN" if the request is not present.
func mGetRequestURI(c echo.Context) string {
	if c.Request() == nil {
		return "UNKNOWN"
	}
	return c.Request().RequestURI
}
