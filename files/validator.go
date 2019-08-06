package files

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"mime/multipart"
	"strings"
)

func Validate(
	fileHeader *multipart.FileHeader,
	file multipart.File,
	submittedFileFieldName string,
	maxUploadFileSizeMb int64,
) (map[string][]string, error) {
	errOutput := map[string][]string{}
	curMime, _, err := mimetype.DetectReader(file)
	if err != nil {
		return errOutput, err
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
		errOutput[submittedFileFieldName] = []string{
			"Not supported image type, supported types are " + SUPPORTED_IMAGE_FORMATS,
		}
	}

	if maxUploadFileSizeMb*1024*1024 < fileHeader.Size {
		errOutput[submittedFileFieldName] = []string{
			fmt.Sprint(
				"The file is too large (%d bytes). Allowed maximum size is %d Mb.",
				fileHeader.Size,
				maxUploadFileSizeMb,
			),
		}
	}

	return errOutput, nil
}
