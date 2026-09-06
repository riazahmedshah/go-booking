package property

import "time"

type PropertyDetailsRaw struct {
	// Property Details
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	SubTitle  *string   `json:"subTitle" db:"sub_title"`
	Price     float64   `json:"price" db:"price"`
	MaxGuests int       `json:"maxGuests" db:"max_guests"`
	Images    []string  `json:"images" db:"images"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	// Host Details (users table)
	HostID   string `json:"hostId" db:"host_id"`
	HostName string `json:"hostName" db:"host_name"`

	// Address Details (addresses table)
	AddressID string  `json:"addressId" db:"address_id"`
	Country   string  `json:"country" db:"country"`
	State     string  `json:"state" db:"state"`
	Pincode   string  `json:"pincode" db:"pincode"`
	City      *string `json:"city" db:"city"`
	Area      string  `json:"area" db:"area"`
}

type Property struct {
	ID        string     `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	SubTitle  *string    `json:"subTitle" db:"sub_title"`
	Price     *float64   `json:"price" db:"price"`
	HostID    string     `json:"hostId" db:"host_id"`
	AddressID *string    `json:"addressId" db:"address_id"`
	MaxGuests *int       `json:"maxGuests" db:"max_guests"`
	CreatedAt *time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt *time.Time `json:"updatedAt" db:"updated_at"`
}

type PropertyImages struct {
	ID         string    `json:"id" db:"id"`
	PropertyID string    `json:"propertyId" db:"property_id"`
	Key        string    `json:"key" db:"key"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

type uploadResult struct {
	ImageID string
	Err     error
}

type PropertyAvailabiliy struct {
	ID          string    `json:"id" db:"id"`
	PropertyID  string    `json:"propertyId" db:"property_id"`
	Date        time.Time `json:"date" db:"date"`
	IsAvailable bool      `json:"isAvailable" db:"is_available"`
	BookingID   *string   `json:"bookingId" db:"booking_id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type DayAvailability struct {
	CalendarDate string `json:"calendarDate"`
	Available    bool   `json:"available"`
}

type MonthAvailability struct {
	Month int               `json:"month"`
	Year  int               `json:"year"`
	Days  []DayAvailability `json:"days"`
}

type Address struct {
	ID         string  `json:"id" db:"id"`
	Country    string  `json:"country" db:"country"`
	State      string  `json:"state" db:"state"`
	Pincode    string  `json:"pincode" db:"pincode"`
	City       *string `json:"city" db:"city"`
	Area       string  `json:"area" db:"area"`
	PropertyID *string `json:"propertyId" db:"property_id"`
	// CreatedAt  *time.Time `json:"createdAt" db:"created_at"`
	// UpdatedAt  *time.Time `json:"updatedAt" db:"updated_at"`
}

type PropertyWithAddress struct {
	Property Property          `json:"property"`
	Address  Address           `json:"address"`
	Images   []*PropertyImages `json:"images"`
}
