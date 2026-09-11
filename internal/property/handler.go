package property

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/riazahmedshah/stayz/internal/lib/utils"
	"github.com/riazahmedshah/stayz/internal/server"
)

type PropertyHandler struct {
	server          *server.Server
	propertyService *PropertyService
}

func NewPropertyHandler(server *server.Server, propertyService *PropertyService) *PropertyHandler {
	return &PropertyHandler{
		server:          server,
		propertyService: propertyService,
	}
}

func (ph *PropertyHandler) CreateProperty(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	var payload CreatePropertyAndAddressPayload

	const maxPayloadSize = (20 << 20) + (10 << 10)
	c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxPayloadSize)

	propertyJSON := c.FormValue("property")
	if propertyJSON == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing property field")
	}
	if err := json.Unmarshal([]byte(propertyJSON), &payload.Property); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid property json")
	}

	addressJSON := c.FormValue("address")
	if addressJSON == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing address field")
	}
	if err := json.Unmarshal([]byte(addressJSON), &payload.Address); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid address json")
	}

	// validating the payload
	if err := c.Validate(&payload.Property); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&payload.Address); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	form, err := c.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse multipart form")
	}

	files := form.File["images"]

	property, err := ph.propertyService.CreateProperty(c.Request().Context(), files, userID, &payload)
	if err != nil {
		return err
	}
	return utils.Success(c, http.StatusCreated, "property created successfully", property)
}

func (ph *PropertyHandler) GetAllProperties(c echo.Context) error {
	properties, err := ph.propertyService.GetAllProperties(c.Request().Context())
	if err != nil {
		return err
	}
	return utils.Success(c, http.StatusOK, "properties fetched successfully", properties)
}

func (ph *PropertyHandler) GetPropertyById(c echo.Context) error {
	propertyID := c.Param("id")
	property, err := ph.propertyService.GetPropertyByID(c.Request().Context(), propertyID)
	if err != nil {
		return err
	}

	return utils.Success(c, http.StatusOK, "property fetched successfully", property)
}

func (ph *PropertyHandler) GetPropertyAvailability(c echo.Context) error {
	propertyID := c.Param("id")
	availability, err := ph.propertyService.GetPropertyAvailability(c.Request().Context(), propertyID)
	if err != nil {
		return err
	}

	return utils.Success(c, http.StatusOK, "property availability fetched successfully", availability)
}

func (ph *PropertyHandler) GetPropertiesByHostID(c echo.Context) error {
	hostID := c.Get("userID").(string)
	properties, err := ph.propertyService.GetPropertiesByHostID(c.Request().Context(), hostID)
	if err != nil {
		return err
	}

	return utils.Success(c, http.StatusOK, "properties fetched successfully", properties)
}
