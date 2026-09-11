package router

import (
	"github.com/labstack/echo/v4"
	Middleware "github.com/labstack/echo/v4/middleware"
	"github.com/riazahmedshah/stayz/internal/handler"
	"github.com/riazahmedshah/stayz/internal/middleware"
	v1 "github.com/riazahmedshah/stayz/internal/router/v1"
	"github.com/riazahmedshah/stayz/internal/server"
	"github.com/riazahmedshah/stayz/internal/validation"
)

func NewRouter(s *server.Server, h *handler.Handler) *echo.Echo {
	middlewares := middleware.NewMiddleware(s)
	router := echo.New()

	router.Use(Middleware.CORSWithConfig(Middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"}, // React/Vite/Next port
		AllowMethods:     []string{echo.GET, echo.POST, echo.PATCH, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	router.HTTPErrorHandler = middleware.ErrMiddleware()
	router.Validator = validation.NewCustomValidator()

	// Register your routes here
	v1Group := router.Group("/api/v1")
	v1.Registerv1Routes(v1Group, h, middlewares)

	return router
}
