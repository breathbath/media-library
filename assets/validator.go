package assets

import (
	"fmt"
	error2 "github.com/breathbath/media-service/error"
	"github.com/gabriel-vasile/mimetype"
	"mime/multipart"
	"strings"
)

func Validate(
	fileHeader *multipart.FileHeader,
	file multipart.File,
	submittedFileFieldName string,
	maxUploadFileSizeMb int64,
) (error2.ValidationErrors, error) {
	validationErrors := error2.NewValidationErrors()
	curMime, _, err := mimetype.DetectReader(file)
	if err != nil {
		return validationErrors, err
	}

	isCurMimeSupported := false
	for _, supportedExt := range strings.Split(SUPPORTED_IMAGE_FORMATS, "|") {
		supportedMime := "image/" + supportedExt
		if curMime == supportedMime {
			isCurMimeSupported = true
			break
		}
	}

	if !isCurMimeSupported {
		validationErrors[submittedFileFieldName] = []string{
			"Not supported image type, supported types are " + SUPPORTED_IMAGE_FORMATS,
		}
	}

	if maxUploadFileSizeMb*1024*1024 < fileHeader.Size {
		validationErrors[submittedFileFieldName] = []string{
			fmt.Sprint(
				"The file is too large (%d bytes). Allowed maximum size is %d Mb.",
				fileHeader.Size,
				maxUploadFileSizeMb,
			),
		}
	}

	return validationErrors, nil
}
