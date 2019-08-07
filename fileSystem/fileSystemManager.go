package fileSystem

import (
	"image"
	"io"
	"net/http"
)

type FileSystemManager interface {
	RemoveNonResizedImage(imgPath ImagePath) error
	IsNonExistingPathError(err error) bool
	IsNonResizedImageDirEmpty(imgPath ImagePath) (bool, error)
	RemoveDir(imgPath ImagePath, isResizedDir bool) error
	CreateNonResizedFileWriter(folderName, imageName string) (io.Writer, error)
	FileExists(imgPath ImagePath, isResized bool) (bool, error)
	CreateFileReader(imgPath ImagePath, isResized bool) (http.File, error)
	OpenNonResizedImage(imgPath ImagePath) (image.Image, error)
	SaveResizedImage(imgPath ImagePath, srcImage image.Image) (http.File, error)
}
