package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/stayz/internal/handler"
	"github.com/riazahmedshah/stayz/internal/middleware"
)

func registerBookingRoutes(r *echo.Group, h *handler.Handler, middlewares *middleware.Middlewares) {
	booking := r.Group("/booking")
	booking.Use(middlewares.Auth.RequireAuth())
	booking.POST("", h.BookingHandler.CreateBooking)
	booking.POST("/:idempotency_key/confirm", h.BookingHandler.ConfirmBooking)
}
