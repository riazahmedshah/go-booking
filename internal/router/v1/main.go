package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/stayz/internal/handler"
	"github.com/riazahmedshah/stayz/internal/middleware"
)

func Registerv1Routes(router *echo.Group, h *handler.Handler, middlewares *middleware.Middlewares) {
	// Register your v1 routes here
	registerUserRoutes(router, h, middlewares)
	registerPropertyRoutes(router, h, middlewares)
	registerBookingRoutes(router, h, middlewares)
}
