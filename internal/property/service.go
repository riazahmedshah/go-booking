package property

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/lib/gcs"
	"github.com/riazahmedshah/go-booking/internal/server"
)

const (
	msgCreatePropertyFailed          = "unexpected error occurred while creating property"
	msgGetAllPropertiesFailed        = "unexpected error occurred while retrieving all properties"
	msgGetPropertyByIDFailed         = "unable to fetch property details"
	msgGetPropertyAvailabilityFailed = "unexpected error occurred while retrieving property availability"
)

type PropertyService struct {
	server       *server.Server
	propertyRepo *PropertyRepository
	gcsClient    *gcs.GCSClient
}

func NewPropertyService(server *server.Server, propertyRepo *PropertyRepository, gcsClient *gcs.GCSClient) *PropertyService {
	return &PropertyService{
		server:       server,
		propertyRepo: propertyRepo,
		gcsClient:    gcsClient,
	}
}

func (ps *PropertyService) CreateProperty(ctx context.Context, hostID string, payload *CreatePropertyAndAddressPayload) (*PropertyWithAddress, error) {
	tx, err := ps.server.DB.Begin(ctx)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}
	defer tx.Rollback(ctx)

	property, err := ps.propertyRepo.Createproperty(ctx, tx, hostID, &payload.Property)
	if err != nil {
		if errors.Is(err, errs.ErrPropertyTitleExists) {
			return nil, err
		}
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}

	payload.Address.PropertyID = property.ID
	address, err := ps.propertyRepo.CreateAddress(ctx, tx, &payload.Address)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}

	propertyWithAddress := &PropertyWithAddress{
		Property: *property,
		Address:  *address,
	}

	return propertyWithAddress, nil
}

func (ps *PropertyService) GetAllProperties(ctx context.Context) ([]*Property, error) {
	properties, err := ps.propertyRepo.GetAllProperties(ctx)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetAllPropertiesFailed, err)
	}

	return properties, nil
}

func (ps *PropertyService) GetPropertyByID(ctx context.Context, propertyID string) (*PropertyDetailsResponse, error) {
	raw, err := ps.propertyRepo.GetPropertyByID(ctx, propertyID)
	if err != nil {
		if errors.Is(err, errs.ErrPropertyNotFound) {
			return nil, err
		}

		return nil, errs.New(http.StatusInternalServerError, msgGetPropertyByIDFailed, err)
	}

	response := &PropertyDetailsResponse{
		ID:        raw.ID,
		Title:     raw.Title,
		SubTitle:  raw.SubTitle,
		Price:     raw.Price,
		MaxGuests: raw.MaxGuests,
		Images:    raw.Images,
		Host: HostResponse{
			ID:   raw.HostID,
			Name: raw.HostName,
		},
		Address: AddressResponse{
			Country: raw.Country,
			State:   raw.State,
			Pincode: raw.Pincode,
			City:    raw.City,
			Area:    raw.Area,
		},
	}

	return response, nil
}

func (ps *PropertyService) GetPropertyAvailability(ctx context.Context, propertyID string) ([]MonthAvailability, error) {
	rows, err := ps.propertyRepo.GetPropertyAvailability(ctx, propertyID)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetPropertyAvailabilityFailed, err)
	}

	// Group by "Year-Month" using a Map
	// Key: "2026-08", Value: pointer to MonthAvailability
	monthMap := make(map[string]*MonthAvailability)
	var result []MonthAvailability

	for _, item := range rows {
		y, m, _ := item.Date.Date()
		mapKey := fmt.Sprintf("%d-%02d", y, int(m))

		if _, exists := monthMap[mapKey]; !exists {
			monthMap[mapKey] = &MonthAvailability{
				Month: int(m),
				Year:  y,
				Days:  []DayAvailability{},
			}

			result = append(result, MonthAvailability{
				Month: int(m),
				Year:  y,
			})
		}

		dayObj := DayAvailability{
			CalendarDate: item.Date.Format("2006-01-02"),
			Available:    item.IsAvailable,
		}

		monthMap[mapKey].Days = append(monthMap[mapKey].Days, dayObj)
	}

	var finalCalendar []MonthAvailability
	for key, mVal := range monthMap {
		_ = key
		finalCalendar = append(finalCalendar, *mVal)
	}

	return finalCalendar, nil
}
