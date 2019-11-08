package fileSystem

import (
	"github.com/breathbath/go_utils/utils/io"
	"github.com/disintegration/imaging"
	"image"
	io2 "io"
	"net/http"
	"os"
	"path/filepath"
)

type LocalFileSystemManager struct {
	AssetsPath string
}

func (lfsm LocalFileSystemManager) IsNonExistingPathError(err error) bool {
	return os.IsNotExist(err)
}

func (lfsm LocalFileSystemManager) RemoveNonResizedImage(imgPath ImagePath) error {
	nonResizedFilePath := imgPath.GetNonResizedImagePath()
	return os.Remove(filepath.Join(lfsm.AssetsPath, nonResizedFilePath))
}

func (lfsm LocalFileSystemManager) CreateNonResizedFileWriter(folderName, imageName string) (io2.WriteCloser, error) {
	err := os.MkdirAll(filepath.Join(lfsm.AssetsPath, folderName), os.ModePerm)
	if err != nil {
		return nil, err
	}

	imgPath := filepath.Join(lfsm.AssetsPath, folderName, imageName)
	io.OutputInfo("", "Will save image under '%s'", filepath.Join(lfsm.AssetsPath, folderName, imageName))
	outfile, err := os.Create(imgPath)
	if err != nil {
		return nil, err
	}

	return outfile, nil
}

func (lfsm LocalFileSystemManager) SaveResizedImage(imgPath ImagePath, srcImage image.Image) (http.File, error) {
	err := os.MkdirAll(filepath.Join(lfsm.AssetsPath, imgPath.GetResizedFolderPath()), os.ModePerm)
	if err != nil {
		return nil, err
	}

	resizedPath := filepath.Join(lfsm.AssetsPath, imgPath.GetResizedImagePath())
	err = imaging.Save(srcImage, resizedPath)
	if err != nil {
		return nil, err
	}

	return os.Open(resizedPath)
}

func (lfsm LocalFileSystemManager) IsImageDirEmpty(imgPath ImagePath, isResized bool) (bool, error) {
	name := filepath.Join(lfsm.AssetsPath, imgPath.GetNonResizedFolderPath())
	if isResized {
		name = filepath.Join(lfsm.AssetsPath, imgPath.GetResizedParentFolderPath())
	}

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

	// read in ONLY one file/subdir
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io2.EOF {
		return true, nil
	}
	return false, err
}

func (lfsm LocalFileSystemManager) FileExists(imgPath ImagePath, isResized bool) (bool, error) {
	filePath := imgPath.GetNonResizedImagePath()
	if isResized {
		filePath = imgPath.GetResizedImagePath()
	}

	filePath = filepath.Join(lfsm.AssetsPath, filePath)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return false, nil
	}

	return true, nil
}

func (lfsm LocalFileSystemManager) CreateFileReader(imgPath ImagePath, isResized bool) (http.File, error) {
	filePath := imgPath.GetNonResizedImagePath()
	if isResized {
		filePath = imgPath.GetResizedImagePath()
	}

	filePath = filepath.Join(lfsm.AssetsPath, filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (lfsm LocalFileSystemManager) OpenNonResizedImage(imgPath ImagePath) (image.Image, error) {
	nonResizedImagePath := filepath.Join(lfsm.AssetsPath, imgPath.FolderName, imgPath.ImageFile)
	_, err := os.Stat(nonResizedImagePath)

	if err != nil {
		return nil, err
	}

	src, err := imaging.Open(nonResizedImagePath)
	if err != nil {
		return nil, err
	}

	return src, nil
}

func (lfsm LocalFileSystemManager) RemoveDir(imgPath ImagePath, isResizedDir, isResizedParentDir bool) error {
	dirToDelete := imgPath.GetNonResizedFolderPath()
	if isResizedDir {
		dirToDelete = imgPath.GetResizedFolderPath()
	}
	if isResizedParentDir {
		dirToDelete = imgPath.GetResizedParentFolderPath()
	}
	return os.RemoveAll(filepath.Join(lfsm.AssetsPath, dirToDelete))
}
