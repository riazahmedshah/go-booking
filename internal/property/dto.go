package property

import "github.com/go-playground/validator/v10"

type CreateAddressPayload struct {
	Country    string  `json:"country" validate:"required"`
	State      string  `json:"state" validate:"required"`
	Pincode    string  `json:"pincode" validate:"required"`
	City       *string `json:"city"`
	Area       string  `json:"area" validate:"required"`
	PropertyID string  `json:"propertyId" validate:"omitempty"`
}

type CreatePropertyPayload struct {
	Title     string   `json:"title" validate:"required,min=1,max=255"`
	SubTitle  *string  `json:"subTitle" validate:"omitempty,max=1000"`
	Price     *float64 `json:"price" validate:"required,min=0"`
	MaxGuests *int     `json:"maxGuests" validate:"omitempty,min=1"`
}

type CreatePropertyAndAddressPayload struct {
	Property CreatePropertyPayload `json:"property" validate:"required"`
	Address  CreateAddressPayload  `json:"address" validate:"required"`
}

func (p *CreatePropertyAndAddressPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// Bug: Review properly.
type UpdatePropertyPayload struct {
	SubTitle  *string `json:"subTitle" validate:"omitempty,max=1000"`
	AddressID *string `json:"addressId" validate:"omitempty"`
	MaxGuests *int    `json:"maxGuests" validate:"omitempty,min=1"`
}

func (p *UpdatePropertyPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type HostResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AddressResponse struct {
	Country string  `json:"country"`
	State   string  `json:"state"`
	Pincode string  `json:"pincode"`
	City    *string `json:"city"`
	Area    string  `json:"area"`
}

type PropertyDetailsResponse struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	SubTitle  *string         `json:"subTitle"`
	Price     float64         `json:"price"`
	MaxGuests int             `json:"maxGuests"`
	Images    []string        `json:"images"`
	Host      HostResponse    `json:"host"`
	Address   AddressResponse `json:"address"`
}

// TODO: other payloads...
