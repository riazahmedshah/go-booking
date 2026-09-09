package utils

import "github.com/labstack/echo/v4"

func Success(c echo.Context, statusCode int, message string, data any) error {
	return c.JSON(statusCode, map[string]any{
		"success": true,
		"data":    data,
		"message": message,
	})
}
