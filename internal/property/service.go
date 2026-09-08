package property

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
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

func allowedMimeTypes(fileHeader *multipart.FileHeader) bool {
	allowedMimeTypes := []string{
		"image/jpeg",
		"image/png",
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return false
	}
	defer srcFile.Close()

	buffer := make([]byte, 512)
	_, err = srcFile.Read(buffer)
	if err != nil {
		return false
	}

	mimeType := http.DetectContentType(buffer)

	for _, allowedType := range allowedMimeTypes {
		if mimeType == allowedType {
			return true
		}
	}
	return false
}

func (ps *PropertyService) CreateProperty(ctx context.Context, files []*multipart.FileHeader, hostID string, payload *CreatePropertyAndAddressPayload) (*PropertyWithAddress, error) {
	if len(files) > 4 {
		return nil, errs.New(http.StatusBadRequest, "maximum 4 files allowed", nil)
	}

	const maxFileSize = 5 << 20 // 5MB in bytes

	for _, file := range files {
		if !allowedMimeTypes(file) {
			return nil, errs.New(http.StatusBadRequest, fmt.Sprintf("file %s has an invalid mime type", file.Filename), nil)
		}
		if file.Size > maxFileSize {
			return nil, errs.New(http.StatusBadRequest, fmt.Sprintf("file %s exceeds 5MB limit", file.Filename), nil)
		}

	}

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

	fileBytes := make([][]byte, 0, len(files))
	keys := make([]string, 0, len(files))

	for _, file := range files {
		randomFileUUID := uuid.Must(uuid.NewV7())
		ext := filepath.Ext(file.Filename)
		key := fmt.Sprintf("%s/%s/%s%s", hostID, property.ID, randomFileUUID.String(), ext)
		keys = append(keys, key)

		f, err := file.Open()
		if err != nil {
			return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
		}
		fileBytes = append(fileBytes, data)
	}
	images, err := ps.propertyRepo.CreatePropertyImages(ctx, tx, property.ID, keys)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgCreatePropertyFailed, err)
	}

	detachedCtx := context.WithoutCancel(ctx)
	go processImageUploads(detachedCtx, ps.gcsClient, ps.propertyRepo, images, fileBytes)

	propertyWithAddress := &PropertyWithAddress{
		Property: *property,
		Address:  *address,
		Images:   images,
	}

	return propertyWithAddress, nil
}

func (ps *PropertyService) GetAllProperties(ctx context.Context) ([]*PopulatedProperty, error) {
	properties, err := ps.propertyRepo.GetAllProperties(ctx)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetAllPropertiesFailed, err)
	}

	return properties, nil
}

func (ps *PropertyService) GetPropertiesByHostID(ctx context.Context, hostID string) ([]*PopulatedProperty, error) {
	properties, err := ps.propertyRepo.GetHostListings(ctx, hostID)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetAllPropertiesFailed, err)
	}

	return properties, nil
}

func (ps *PropertyService) GetPropertyByID(ctx context.Context, propertyID string) (*PopulatedPropertyWithHost, error) {
	property, err := ps.propertyRepo.GetPropertyByID(ctx, propertyID)
	if err != nil {
		if errors.Is(err, errs.ErrPropertyNotFound) {
			return nil, err
		}

		return nil, errs.New(http.StatusInternalServerError, msgGetPropertyByIDFailed, err)
	}

	return property, nil
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

func processImageUploads(ctx context.Context, gcsClient *gcs.GCSClient, propertyRepo *PropertyRepository, images []*PropertyImages, files [][]byte) {
	resultChan := make(chan uploadResult, len(images))
	var wg sync.WaitGroup

	for i, image := range images {
		wg.Add(1)
		go func(img *PropertyImages, fileData []byte) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("panic during image upload", "imageId", img.ID, "error", r)
					resultChan <- uploadResult{ImageID: img.ID, Err: fmt.Errorf("panic: %v", r)}
				}
			}()

			err := gcsClient.UploadFile(ctx, img.Key, fileData)
			resultChan <- uploadResult{ImageID: img.ID, Err: err}
		}(image, files[i])
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		status := "active"
		if result.Err != nil {
			status = "failed"
			slog.Error("image upload failed", "imageId", result.ImageID, "error", result.Err)
		}
		if err := propertyRepo.UpdateImageStatus(ctx, result.ImageID, status); err != nil {
			slog.Error("failed to update image status", "imageId", result.ImageID, "error", err)
		}
	}
}
