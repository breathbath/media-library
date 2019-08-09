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
	maxUploadFileSizeMb float64,
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
			fmt.Sprintf(
				"Not supported image type '%s', supported types are %s",
				curMime,
				SUPPORTED_IMAGE_FORMATS,
			),
		}
	}

	if maxUploadFileSizeMb*1024.0*1024.0 < float64(fileHeader.Size) {
		validationErrors[submittedFileFieldName] = []string{
			fmt.Sprintf(
				"The file is too large (%d bytes). Allowed maximum size is %v Mb",
				fileHeader.Size,
				maxUploadFileSizeMb,
			),
		}
	}

	return validationErrors, nil
}
