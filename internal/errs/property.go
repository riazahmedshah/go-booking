package errs

import "net/http"

var (
	ErrPropertyTitleExists = &AppError{
		StatusCode: http.StatusConflict,
		Code:       "PROPERTY_TITLE_EXISTS",
		Message:    "use different title for property, this title already exists",
		Op:         "createProperty.titleCheck",
	}

	ErrPropertyNotFound = &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "PROPERTY_NOT_FOUND",
		Message:    "property not found",
		Op:         "getProperty.idCheck",
	}

	ErrBadUpdateRequest = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_UPDATE_REQUEST",
		Message:    "invalid update fields",
		Op:         "updateProperty.fieldsCheck",
	}

	ErrImageNotFound = &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "IMAGE_NOT_FOUND",
		Message:    "image not found",
		Op:         "getImage.idCheck",
	}

	ErrInvalidImageStatus = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_IMAGE_STATUS",
		Message:    "invalid image status",
		Op:         "updateImage.statusCheck",
	}

	//
	ErrLimitFilesExceeded = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "LIMIT_FILES_EXCEEDED",
		Message:    "maximum 4 files allowed",
		Op:         "createProperty.filesCheck",
	}

	ErrInvalidFileType = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_FILE_TYPE",
		Message:    "invalid file type, only JPEG and PNG are allowed",
		Op:         "createProperty.fileTypeCheck",
	}
	ErrFileTooLarge = &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "FILE_TOO_LARGE",
		Message:    "file size exceeds the maximum limit of 5MB",
		Op:         "createProperty.fileSizeCheck",
	}
)
