package user

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/go-booking/internal/lib/utils"
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
	return utils.Success(c, http.StatusOK, "otp send successfully", nil)
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

	return utils.Success(c, http.StatusOK, "otp verified successfully", result)
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

	cookie := new(http.Cookie) // #nosec G124
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = uh.server.Config.Env == "production"
	cookie.Path = "/"

	c.SetCookie(cookie)

	return utils.Success(c, http.StatusCreated, "user created successfully", nil)
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

	cookie := new(http.Cookie) // #nosec G124
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = uh.server.Config.Env == "production"
	cookie.Path = "/"

	c.SetCookie(cookie)
	return utils.Success(c, http.StatusOK, "logged in successfully", nil)
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

	cookie := new(http.Cookie) // #nosec G124
	cookie.Name = "sid"
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = uh.server.Config.Env == "production"
	cookie.Path = "/"

	c.SetCookie(cookie)
	return utils.Success(c, http.StatusOK, "logged in with google successfully", nil)
}

func (uh *UserHandler) GetCurrentUser(c echo.Context) error {
	userID := c.Get("userID").(string)
	user, err := uh.userService.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return utils.Success(c, http.StatusOK, "current user fetched successfully", user)
}

func (uh *UserHandler) UpdateRole(c echo.Context) error {
	userID := c.Get("userID").(string)
	err := uh.userService.UpdateRole(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return utils.Success(c, http.StatusOK, "user role updated successfully", nil)
}
