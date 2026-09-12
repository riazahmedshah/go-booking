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
	"github.com/riazahmedshah/stayz/internal/errs"
	"github.com/riazahmedshah/stayz/internal/lib/gcs"
	"github.com/riazahmedshah/stayz/internal/server"
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
		return nil, errs.ErrLimitFilesExceeded
	}

	const maxFileSize = 5 << 20 // 5MB in bytes

	for _, file := range files {
		if !allowedMimeTypes(file) {
			return nil, errs.ErrInvalidFileType
		}
		if file.Size > maxFileSize {
			return nil, errs.ErrFileTooLarge
		}

	}

	tx, err := ps.server.DB.Begin(ctx)
	if err != nil {
		return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.beginTx", err)
	}
	defer tx.Rollback(ctx)

	property, err := ps.propertyRepo.Createproperty(ctx, tx, hostID, &payload.Property)
	if err != nil {
		if errors.Is(err, errs.ErrPropertyTitleExists) {
			return nil, err
		}
		return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.create", err)
	}

	payload.Address.PropertyID = property.ID
	address, err := ps.propertyRepo.CreateAddress(ctx, tx, &payload.Address)
	if err != nil {
		return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.createAddress", err)
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
			return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.openFile", err)
		}
		data, err := io.ReadAll(f)
		if cerr := f.Close(); cerr != nil {
			slog.Info("failed to close file", "err", cerr)
		}
		if err != nil {
			return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.readFile", err)
		}
		fileBytes = append(fileBytes, data)
	}
	images, err := ps.propertyRepo.CreatePropertyImages(ctx, tx, property.ID, keys)
	if err != nil {
		return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.createImages", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.Internal(msgCreatePropertyFailed, "createProperty.commitTx", err)
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
		return nil, errs.Internal(msgGetAllPropertiesFailed, "getAllProperties", err)
	}

	filteredProperties := make([]*PopulatedProperty, 0, len(properties))
	status := "active"
	for _, property := range properties {
		filteredPropertyImages := make([]PropertyImage, 0)
		for _, img := range property.Images {
			if *img.Status == status {
				filteredPropertyImages = append(filteredPropertyImages, img)
			}
		}
		if len(filteredPropertyImages) > 0 {
			property.Images = filteredPropertyImages
			filteredProperties = append(filteredProperties, property)
		}
	}

	return filteredProperties, nil
}

func (ps *PropertyService) GetPropertiesByHostID(ctx context.Context, hostID string) ([]*PopulatedProperty, error) {
	properties, err := ps.propertyRepo.GetHostListings(ctx, hostID)
	if err != nil {
		return nil, errs.Internal(msgGetAllPropertiesFailed, "getPropertiesByHostID", err)
	}

	return properties, nil
}

func (ps *PropertyService) GetPropertyByID(ctx context.Context, propertyID string) (*PopulatedPropertyWithHost, error) {
	property, err := ps.propertyRepo.GetPropertyByID(ctx, propertyID)
	if err != nil {
		if errors.Is(err, errs.ErrPropertyNotFound) {
			return nil, err
		}

		return nil, errs.Internal(msgGetPropertyByIDFailed, "getPropertyByID", err)
	}

	return property, nil
}

func (ps *PropertyService) GetPropertyAvailability(ctx context.Context, propertyID string) ([]MonthAvailability, error) {
	rows, err := ps.propertyRepo.GetPropertyAvailability(ctx, propertyID)
	if err != nil {
		return nil, errs.Internal(msgGetPropertyAvailabilityFailed, "getPropertyAvailability", err)
	}

	// Group by "Year-Month" using a Map
	// Key: "2026-08", Value: pointer to MonthAvailability
	monthMap := make(map[string]*MonthAvailability)
	// var result []MonthAvailability
	// var order []string // preserves first-seen order of "YYYY-MM" keys

	for _, item := range rows {
		y, m, _ := item.Date.Date()
		mapKey := fmt.Sprintf("%d-%02d", y, int(m))

		if _, exists := monthMap[mapKey]; !exists {
			monthMap[mapKey] = &MonthAvailability{
				Month: int(m),
				Year:  y,
				Days:  []DayAvailability{},
			}

			// order = append(order, mapKey)
			// result = append(result, MonthAvailability{
			// 	Month: int(m),
			// 	Year:  y,
			// })
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

func (ps *PropertyService) SearchProperties(ctx context.Context, searchPayload *SearchPropertyPayload) ([]*PopulatedProperty, error) {
	properties, err := ps.propertyRepo.Search(ctx, searchPayload)
	if err != nil {
		return nil, errs.Internal("unexpected error occurred while searching properties", "searchProperties", err)
	}

	filteredProperties := make([]*PopulatedProperty, 0, len(properties))
	status := "active"
	for _, property := range properties {
		filteredPropertyImages := make([]PropertyImage, 0)
		for _, img := range property.Images {
			if *img.Status == status {
				filteredPropertyImages = append(filteredPropertyImages, img)
			}
		}
		if len(filteredPropertyImages) > 0 {
			property.Images = filteredPropertyImages
			filteredProperties = append(filteredProperties, property)
		}
	}

	return filteredProperties, nil
}
