package booking

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/go-booking/internal/server"
)

type BookingHandler struct {
	server         *server.Server
	bookingService *BookingService
}

func NewBookingHandler(server *server.Server, bookingService *BookingService) *BookingHandler {
	return &BookingHandler{
		server:         server,
		bookingService: bookingService,
	}
}

func (bh *BookingHandler) CreateBooking(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	var payload CreateBookingPayload

	if err := c.Bind(&payload); err != nil {
		slog.Error("failed to bind request payload", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	idempotencyKey, err := bh.bookingService.CreateBooking(c.Request().Context(), userID, &payload)
	if err != nil {
		return err
	}

	return c.JSON(201, map[string]string{"idempotency_key": idempotencyKey.(string)})
}

func (bh *BookingHandler) ConfirmBooking(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	idempotencyKey := c.Param("idempotency_key")

	booking, err := bh.bookingService.ConfirmBooking(c.Request().Context(), idempotencyKey, userID)
	if err != nil {
		slog.Error("failed to confirm booking", "error", err)
		return err
	}

	return c.JSON(200, booking)
}
