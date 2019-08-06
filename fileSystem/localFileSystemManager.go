package fileSystem

import (
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-service/assets"
	io2 "io"
	"os"
	"path/filepath"
)

type LocalFileSystemManager struct {
	AssetsPath string
}

func (lfsm LocalFileSystemManager) IsNonExistingPathError(err error) bool {
	return os.IsNotExist(err)
}

func (lfsm LocalFileSystemManager) RemoveNonResizedImage(imgPath assets.ImagePath) error {
	nonResizedFilePath := imgPath.GetNonResizedImagePath()
	return os.Remove(nonResizedFilePath)
}

func (lfsm LocalFileSystemManager) CreateNonResizedFileWriter(folderName, imageName string) (io2.Writer, error) {
	err := os.MkdirAll(filepath.Join(lfsm.AssetsPath, folderName), os.ModePerm)
	if err != nil {
		return nil, err
	}

	outfile, err := os.Create(filepath.Join(lfsm.AssetsPath, folderName, imageName))
	if err != nil {
		return nil, err
	}

	return outfile, nil
}

func (lfsm LocalFileSystemManager) IsNonResizedImageDirEmpty(imgPath assets.ImagePath) (bool, error) {
	name := imgPath.GetNonResizedFolderPath()

	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			io.OutputError(err, "", "Failed to close directory '%s'", name)
		}
	}()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io2.EOF {
		return true, nil
	}
	return false, err
}

func (lfsm LocalFileSystemManager) RemoveDir(imgPath assets.ImagePath, isResizedDir bool) error {
	dirToDelete := imgPath.GetNonResizedFolderPath()
	if isResizedDir {
		dirToDelete = imgPath.GetResizedFolderPath()
	}
	return os.RemoveAll(dirToDelete)
}
