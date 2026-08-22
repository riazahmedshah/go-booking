package user

import (
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

func (uh *UserHandler) CreateUser(c echo.Context) error {
	var userPayload CreateUserPayload
	if err := c.Bind(&userPayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}
	err := uh.userService.CreateUser(c.Request().Context(), &userPayload)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "user created successfully"})
}

func (uh *UserHandler) Login(c echo.Context) error {
	var loginPayload LoginPayload
	if err := c.Bind(&loginPayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	token, err := uh.userService.Login(c.Request().Context(), &loginPayload)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "access_token"
	cookie.Value = token
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.HttpOnly = true
	cookie.Secure = false // Ensures cookie is only sent over HTTPS (Set to false ONLY in local dev if not using HTTPS)
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Path = "/"

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, echo.Map{"message": "logged in successful"})
}

func (uh *UserHandler) GetCurrentUser(c echo.Context) error {
	userID := c.Get("userID").(string)
	user, err := uh.userService.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}
