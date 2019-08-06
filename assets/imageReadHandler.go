package assets

import (
	"github.com/breathbath/media-service/fileSystem"
	"github.com/disintegration/imaging"
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
	if !imagePath.isValid {
		return nil, nfs.createNonExistsError(path)
	}

	if imagePath.rawResizedFolder != "" {
		return nfs.handleResizedImage(imagePath)
	}

	return nfs.handleNonResizedImage(imagePath)
}

func (nfs ImageReadHandler) createNonExistsError(path string) *os.PathError {
	return &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (nfs ImageReadHandler) handleResizedImage(imagePath ImagePath) (http.File, error) {
	resizedPath := imagePath.GetResizedImagePath()
	info, err := os.Stat(resizedPath)
	if os.IsNotExist(err) {
		f, err := nfs.generateResizedImage(imagePath)
		return f, err
	}

	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nfs.createNonExistsError(resizedPath)
	}

	f, err := os.Open(resizedPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (nfs ImageReadHandler) generateResizedImage(imagePath ImagePath) (http.File, error) {
	resizedFolder := imagePath.GetResizedFolderPath()
	resizedPath := imagePath.GetResizedImagePath()

	err := os.MkdirAll(resizedFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	nonResizedImagePath := imagePath.GetNonResizedImagePath()
	_, err = os.Stat(nonResizedImagePath)

	if err != nil {
		return nil, err
	}

	src, err := imaging.Open(nonResizedImagePath)
	if err != nil {
		return nil, err
	}

	src = imaging.Fill(src, imagePath.width, imagePath.height, imaging.Center, imaging.Lanczos)

	err = imaging.Save(src, resizedPath)
	if err != nil {
		return nil, err
	}

	return os.Open(resizedPath)
}

func (nfs ImageReadHandler) handleNonResizedImage(imagePath ImagePath) (http.File, error) {
	fullImagePath := imagePath.GetNonResizedImagePath()

	info, err := os.Stat(fullImagePath)
	if os.IsNotExist(err) {
		return nil, nfs.createNonExistsError(fullImagePath)
	}
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nfs.createNonExistsError(fullImagePath)
	}

	f, err := os.Open(fullImagePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}
