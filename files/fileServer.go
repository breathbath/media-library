package files

import (
	"github.com/disintegration/imaging"
	"net/http"
	"os"
	"path/filepath"
)

type fileSystemManager struct {
	fs         http.FileSystem
	assetsPath string
}

func NewFileServer(assetsPath string) http.Handler {
	return http.FileServer(fileSystemManager{http.Dir(assetsPath), assetsPath})
}

func (nfs fileSystemManager) Open(path string) (http.File, error) {
	imagePath := parseImagePath(path)
	if !imagePath.isValid {
		return nil, nfs.createNonExistsError(path)
	}

	if imagePath.rawResizedFolder != "" {
		return nfs.handleResizedImage(imagePath)
	}

	return nfs.handleNonResizedImage(imagePath)
}

func (nfs fileSystemManager) createNonExistsError(path string) *os.PathError {
	return &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (nfs fileSystemManager) handleResizedImage(imagePath ImagePath) (http.File, error) {
	resizedFolder := filepath.Join(
		nfs.assetsPath,
		"cache",
		"resized",
		imagePath.folderName,
		imagePath.imageName,
	)

	resizedImage := filepath.Join(
		imagePath.rawResizedFolder,
		".",
		imagePath.imageExt,
	)

	resizedPath := filepath.Join(resizedFolder, resizedImage)

	info, err := os.Stat(resizedPath)
	if os.IsNotExist(err) {
		f, err := nfs.generateResizedImage(resizedFolder, resizedImage, imagePath)
		return f, err
	}

	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nfs.createNonExistsError(resizedPath)
	}

	f, err := nfs.fs.Open(resizedPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (nfs fileSystemManager) generateResizedImage(resizedFolder, resizedImageName string, imagePath ImagePath) (http.File, error) {
	resizedPath := filepath.Join(resizedFolder, resizedImageName)

	err := os.MkdirAll(resizedFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	nonResizedImagePath := filepath.Join(nfs.assetsPath, imagePath.folderName, imagePath.imageFile)
	_, err = os.Stat(nonResizedImagePath)

	if err != nil {
		return nil, err
	}

	src, err := imaging.Open(nonResizedImagePath)
	if err != nil {
		return nil, err
	}

	src = imaging.Resize(src, imagePath.width, imagePath.height, imaging.Lanczos)

	err = imaging.Save(src, resizedPath)
	if err != nil {
		return nil, err
	}

	return nfs.fs.Open(resizedPath)
}

func (nfs fileSystemManager) handleNonResizedImage(imagePath ImagePath) (http.File, error) {
	fullImagePath := filepath.Join(nfs.assetsPath, imagePath.folderName, imagePath.imageFile)

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

	f, err := nfs.fs.Open(fullImagePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}
