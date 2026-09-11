package handler

import (
	"github.com/riazahmedshah/stayz/internal/booking"
	"github.com/riazahmedshah/stayz/internal/property"
	"github.com/riazahmedshah/stayz/internal/server"
	"github.com/riazahmedshah/stayz/internal/service"
	"github.com/riazahmedshah/stayz/internal/user"
)

type Handler struct {
	UserHandler     *user.UserHandler
	PropertyHandler *property.PropertyHandler
	BookingHandler  *booking.BookingHandler
}

func NewHandler(server *server.Server, service *service.Service) *Handler {
	userHandler := user.NewUserHandler(server, service.UserService)
	propertyHandler := property.NewPropertyHandler(server, service.PropertyService)
	bookingHandler := booking.NewBookingHandler(server, service.BookingService)
	return &Handler{
		UserHandler:     userHandler,
		PropertyHandler: propertyHandler,
		BookingHandler:  bookingHandler,
	}
}
