package assets

import (
	"github.com/breathbath/media-service/fileSystem"
	"github.com/disintegration/imaging"
	"math"
	"net/http"
	"os"
)

type ImageReadHandler struct {
	fileSystemManager fileSystem.FileSystemManager
}

func NewImageReadHandler(fileSystemManager fileSystem.FileSystemManager) ImageReadHandler {
	return ImageReadHandler{fileSystemManager: fileSystemManager}
}

func (nfs ImageReadHandler) Open(path string) (http.File, error) {
	imagePath := parseImagePath(path)
	if !imagePath.IsValid {
		return nil, nfs.createNonExistsError(path)
	}

	if imagePath.RawResizedFolder != "" {
		return nfs.handleResizedImage(imagePath)
	}

	return nfs.handleNonResizedImage(imagePath)
}

func (nfs ImageReadHandler) createNonExistsError(path string) *os.PathError {
	return &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (nfs ImageReadHandler) handleResizedImage(imagePath fileSystem.ImagePath) (http.File, error) {
	fileExists, err := nfs.fileSystemManager.FileExists(imagePath, true)
	if err != nil {
		return nil, err
	}

	if !fileExists {
		f, err := nfs.generateResizedImage(imagePath)
		return f, err
	}

	return nfs.fileSystemManager.CreateFileReader(imagePath, true)
}

func (nfs ImageReadHandler) generateResizedImage(imagePath fileSystem.ImagePath) (http.File, error) {
	srcImage, err := nfs.fileSystemManager.OpenNonResizedImage(imagePath)
	if err != nil {
		return nil, err
	}

	if imagePath.Width == 0 {
		srcBound := srcImage.Bounds()
		srcW := srcBound.Dx()
		srcH := srcBound.Dy()
		tmpW := float64(imagePath.Height) * float64(srcW) / float64(srcH)
		imagePath.Width = int(math.Max(1.0, math.Floor(tmpW+0.5)))
	}
	if imagePath.Height == 0 {
		srcBound := srcImage.Bounds()
		srcW := srcBound.Dx()
		srcH := srcBound.Dy()
		tmpH := float64(imagePath.Width) * float64(srcH) / float64(srcW)
		imagePath.Height = int(math.Max(1.0, math.Floor(tmpH+0.5)))
	}

	targetImg := imaging.Thumbnail(srcImage, imagePath.Width, imagePath.Height, imaging.Lanczos)

	file, err := nfs.fileSystemManager.SaveResizedImage(imagePath, targetImg)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (nfs ImageReadHandler) handleNonResizedImage(imagePath fileSystem.ImagePath) (http.File, error) {
	fileExists, err := nfs.fileSystemManager.FileExists(imagePath, false)
	if err != nil {
		return nil, err
	}

	if !fileExists {
		return nil, nfs.createNonExistsError(imagePath.ImageFile)
	}

	return nfs.fileSystemManager.CreateFileReader(imagePath, false)
}
