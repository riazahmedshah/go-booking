package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/go-booking/internal/handler"
	"github.com/riazahmedshah/go-booking/internal/middleware"
)

func registerUserRoutes(r *echo.Group, h *handler.Handler, middlewares *middleware.Middlewares) {
	auth := r.Group("/auth")
	auth.POST("/otp", h.UserHandler.SendOTP)
	auth.POST("/otp/verify", h.UserHandler.VerifyOTP)
	auth.POST("/register", h.UserHandler.CreateUser)
	auth.POST("/login", h.UserHandler.Login)
	auth.GET("/me", h.UserHandler.GetCurrentUser, middlewares.Auth.RequireAuth())
}
