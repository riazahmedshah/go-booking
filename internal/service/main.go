package service

import (
	"fmt"

	"github.com/riazahmedshah/stayz/internal/booking"
	"github.com/riazahmedshah/stayz/internal/lib/gcs"
	"github.com/riazahmedshah/stayz/internal/notification"
	"github.com/riazahmedshah/stayz/internal/property"
	"github.com/riazahmedshah/stayz/internal/repository"
	"github.com/riazahmedshah/stayz/internal/server"
	"github.com/riazahmedshah/stayz/internal/user"
)

type Service struct {
	UserService     *user.UserService
	PropertyService *property.PropertyService
	BookingService  *booking.BookingService
}

func NewService(server *server.Server, repository *repository.Repositories, noti *notification.NotificationService) (*Service, error) {
	gcsClient, err := gcs.NewGCSClient(server.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	userService := user.NewUserService(server, repository.UserRepository, noti)
	propertyService := property.NewPropertyService(server, repository.PropertyRepository, gcsClient)
	bookingService := booking.NewBookingService(server, repository.BookingRepository, noti)
	return &Service{
		UserService:     userService,
		PropertyService: propertyService,
		BookingService:  bookingService,
	}, nil
}
