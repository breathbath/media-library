package fileSystem

import (
	"github.com/breathbath/media-service/assets"
	"io"
)

type FileSystemManager interface {
	RemoveNonResizedImage(imgPath assets.ImagePath) error
	IsNonExistingPathError(err error) bool
	IsNonResizedImageDirEmpty(imgPath assets.ImagePath) (bool, error)
	RemoveDir(imgPath assets.ImagePath, isResizedDir bool) error
	CreateNonResizedFileWriter(folderName, imageName string) (io.Writer, error)
}
