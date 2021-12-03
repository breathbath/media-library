package filesystem

import (
	"image"
	"io"
	"net/http"
)

type Manager interface {
	RemoveNonResizedImage(imgPath *ImagePath) error
	IsNonExistingPathError(err error) bool
	IsImageDirEmpty(imgPath *ImagePath, isResized bool) (bool, error)
	RemoveDir(imgPath *ImagePath, isResizedDir, isResizedParentDir bool) error
	CreateNonResizedFileWriter(folderName, imageName string) (io.WriteCloser, error)
	FileExists(imgPath *ImagePath, isResized bool) (bool, error)
	CreateFileReader(imgPath *ImagePath, isResized bool) (http.File, error)
	OpenNonResizedImage(imgPath *ImagePath) (image.Image, error)
	SaveResizedImage(imgPath *ImagePath, srcImage image.Image) (http.File, error)
}
