package user

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/go-booking/internal/server"
)

type UserHandler struct {
	server      *server.Server
	userService *UserService
}

func NewUserHandler(server *server.Server, us *UserService) *UserHandler {
	return &UserHandler{
		server:      server,
		userService: us,
	}
}

func (uh *UserHandler) SendOTP(c echo.Context) error {
	var payload SendOTPPayload
	if err := c.Bind(&payload); err != nil {
		slog.Error("invalid payload", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := c.Validate(&payload); err != nil {
		return err
	}

	err := uh.userService.SendOTP(c.Request().Context(), payload.Email)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "otp send successfully"})
}

func (uh *UserHandler) VerifyOTP(c echo.Context) error {
	var payload VerifyOTPPayload
	if err := c.Bind(&payload); err != nil {
		slog.Error("invalid payload", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := c.Validate(&payload); err != nil {
		return err
	}

	result, err := uh.userService.VerifyOTP(c.Request().Context(), payload.Email, payload.OTP)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "success",
		"data":    result,
	})
}

func (uh *UserHandler) CreateUser(c echo.Context) error {
	var userPayload CreateUserPayload
	if err := c.Bind(&userPayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}
	sid, err := uh.userService.Register(c.Request().Context(), &userPayload)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.Secure = false // Ensures cookie is only sent over HTTPS (Set to false ONLY in local dev if not using HTTPS)
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Path = "/"

	c.SetCookie(cookie)

	return c.JSON(http.StatusCreated, map[string]string{"message": "user created successfully"})
}

func (uh *UserHandler) Login(c echo.Context) error {
	var loginPayload LoginPayload
	if err := c.Bind(&loginPayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := c.Validate(&loginPayload); err != nil {
		return err
	}

	sid, err := uh.userService.Login(c.Request().Context(), &loginPayload)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.Secure = false // Ensures cookie is only sent over HTTPS (Set to false ONLY in local dev if not using HTTPS)
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Path = "/"

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, echo.Map{"message": "logged in successful"})
}

func (uh *UserHandler) LoginWithGoogle(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing authorization code query parameter",
		})
	}
	sid, err := uh.userService.LoginWithGoogle(c.Request().Context(), code)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.Secure = false // Ensures cookie is only sent over HTTPS (Set to false ONLY in local dev if not using HTTPS)
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Path = "/"

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, echo.Map{"message": "logged in with google successfully"})
}

func (uh *UserHandler) GetCurrentUser(c echo.Context) error {
	userID := c.Get("userID").(string)
	user, err := uh.userService.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}
