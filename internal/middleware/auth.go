package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/go-booking/internal/server"
	"github.com/riazahmedshah/go-booking/internal/user"
)

const (
	UserIDKey = "userID"
	RoleKey   = "userRole"
)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(server *server.Server) *AuthMiddleware {
	return &AuthMiddleware{server: server}
}

func (auth *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("sid")
			if err != nil {
				if err == http.ErrNoCookie {
					slog.Error("cookie not found")
					return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
				}
				return echo.NewHTTPError(http.StatusBadRequest, "bad request")
			}

			sessionID := cookie.Value

			cmd := auth.server.RedisClient.B().Get().Key("session:" + sessionID).Build()
			res, err := auth.server.RedisClient.Do(c.Request().Context(), cmd).AsBytes()
			if err != nil {
				slog.Error("error occurred while fetching session data", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			var sessionData user.SessionData
			if err := json.Unmarshal(res, &sessionData); err != nil {
				slog.Error("error occurred while unmarshalling session data", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}

			c.Set(UserIDKey, sessionData.UserID)
			c.Set(RoleKey, sessionData.Role)

			return next(c)
		}
	}
}

func (auth *AuthMiddleware) RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			slog.Info("checking user role")
			token, ok := c.Get("user").(*jwt.Token)
			if !ok || token == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			userRole, ok := claims["role"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing role in token claims")
			}

			for _, role := range roles {
				if userRole == role {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "you do not have the required permissions to access this resource")
		}
	}
}
