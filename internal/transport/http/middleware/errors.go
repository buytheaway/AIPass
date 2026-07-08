package middleware

import (
	"errors"
	"net/http"

	"github.com/aipass/aipass/internal/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func HTTPErrorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		message := "internal_error"

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			status = httpErr.Code
			message = toMessage(httpErr.Message)
		}
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
			message = "not_found"
		}
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusUnauthorized
			message = "unauthorized"
		}
		if errors.Is(err, service.ErrForbidden) {
			status = http.StatusForbidden
			message = "forbidden"
		}
		if errors.Is(err, service.ErrInvalidInput) {
			status = http.StatusBadRequest
			message = "invalid_input"
		}
		if errors.Is(err, service.ErrConflict) {
			status = http.StatusConflict
			message = "conflict"
		}

		if status >= 500 {
			log.Error("request failed", zap.Error(err))
		}
		_ = c.JSON(status, ErrorResponse{Error: message})
	}
}

func toMessage(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return "request_failed"
}
