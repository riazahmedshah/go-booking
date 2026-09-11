package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/stayz/internal/errs"
)

func ErrMiddleware() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		if appErr, ok := errors.AsType[*errs.AppError](err); ok {
			if appErr.Err != nil {
				slog.Error("internal system error",
					"op", appErr.Op,
					"code", appErr.Code,
					"user_msg", appErr.Message,
					"internal_err", appErr.Err,
					"path", c.Path(),
					"method", c.Request().Method,
				)
			} else {
				slog.Warn("service error",
					"op", appErr.Op,
					"code", appErr.Code,
					"message", appErr.Message,
					"path", c.Path(),
				)
			}

			_ = c.JSON(appErr.StatusCode, map[string]any{
				"success": false,
				"error": map[string]string{
					"code":    appErr.Code,
					"message": appErr.Message,
				},
			})
			return
		}

		if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
			slog.Warn("echo HTTP error", "code", echoErr.Code, "msg", echoErr.Message)
			_ = c.JSON(echoErr.Code, map[string]any{
				"success": false,
				"error": map[string]string{
					"code":    "HTTP_ERROR",
					"message": fmt.Sprintf("%v", echoErr.Message),
				},
			})
			return
		}

		slog.Error("unhandled critical server error", "error", err, "path", c.Path())
		_ = c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "An unexpected server error occurred. Please try again later.",
			},
		})
	}
}
