package assets

import (
	"fmt"
	"mime/multipart"
	"strings"

	error2 "github.com/breathbath/media-library/error"
)

func Validate(
	fileHeader *multipart.FileHeader,
	curMime, submittedFileFieldName string,
	maxUploadFileSizeMb float64,
) (error2.ValidationErrors, error) {
	validationErrors := error2.NewValidationErrors()

	isCurMimeSupported := false
	for _, supportedExt := range strings.Split(SupportedImageFormats, "|") {
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
				SupportedImageFormats,
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
