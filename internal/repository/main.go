package repository

import (
	"github.com/riazahmedshah/stayz/internal/booking"
	"github.com/riazahmedshah/stayz/internal/property"
	"github.com/riazahmedshah/stayz/internal/server"
	"github.com/riazahmedshah/stayz/internal/user"
)

type Repositories struct {
	UserRepository     *user.UserRepository
	PropertyRepository *property.PropertyRepository
	BookingRepository  *booking.BookingRepository
}

func NewRepositories(s *server.Server) *Repositories {
	userRepo := user.NewUserRepository(s)
	propertyRepo := property.NewPropertyRepository(s)
	bookingRepo := booking.NewBookingRepository(s)
	return &Repositories{
		UserRepository:     userRepo,
		PropertyRepository: propertyRepo,
		BookingRepository:  bookingRepo,
	}
}
