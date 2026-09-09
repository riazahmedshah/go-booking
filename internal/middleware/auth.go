package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
	"github.com/riazahmedshah/go-booking/internal/errs"
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
				if errors.Is(err, http.ErrNoCookie) {
					slog.Warn("request without session cookie", "path", c.Path())
					return errs.ErrUnauthorized
				}
				return errs.Internal("failed to read session cookie", "requireAuth.readCookie", err)
			}

			sessionID := cookie.Value

			cmd := auth.server.RedisClient.B().Get().Key("session:" + sessionID).Build()
			res, err := auth.server.RedisClient.Do(c.Request().Context(), cmd).AsBytes()
			if err != nil {
				// Redis nil = session expired/not found, ye "unauthorized" hai, "internal error" nahi
				if rueidis.IsRedisNil(err) {
					slog.Warn("session not found or expired", "path", c.Path())
					return errs.ErrUnauthorized
				}
				return errs.Internal("failed to fetch session data", "requireAuth.redisGet", err)
			}

			var sessionData user.SessionData
			if err := json.Unmarshal(res, &sessionData); err != nil {
				return errs.Internal("failed to parse session data", "requireAuth.unmarshal", err)
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
			userRole, ok := c.Get(RoleKey).(string)
			if !ok {
				return errs.Internal("role missing from context, RequireAuth may not have run", "requireRole.contextCheck", nil)
			}

			if slices.Contains(roles, userRole) {
				return next(c)
			}

			slog.Warn("forbidden access attempt", "role", userRole, "path", c.Path())
			return errs.ErrForbidden
		}
	}
}
