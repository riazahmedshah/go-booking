package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/stayz/internal/handler"
	"github.com/riazahmedshah/stayz/internal/middleware"
)

func registerPropertyRoutes(r *echo.Group, h *handler.Handler, middlewares *middleware.Middlewares) {
	r.GET("/property", h.PropertyHandler.GetAllProperties)
	r.GET("/property/:id", h.PropertyHandler.GetPropertyById)
	r.GET("/property/:id/availability", h.PropertyHandler.GetPropertyAvailability)
	r.GET("/property/search", h.PropertyHandler.SearchProperties)
	property := r.Group("/property")
	property.Use(middlewares.Auth.RequireAuth())
	property.GET("/property/host/hostings", h.PropertyHandler.GetPropertiesByHostID)
	property.POST("", h.PropertyHandler.CreateProperty, middlewares.Auth.RequireRole("host"))
}
