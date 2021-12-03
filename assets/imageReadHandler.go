package assets

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/media-library/filesystem"
	"github.com/disintegration/imaging"
)

type ImageReadHandler struct {
	fileSystemManager filesystem.Manager
	proxyURL          string
}

func NewImageReadHandler(fileSystemManager filesystem.Manager) ImageReadHandler {
	proxyURL := env.ReadEnv("PROXY_URL", "")
	if proxyURL != "" {
		urlPrefix := env.ReadEnv("URL_PREFIX", "/media/images")
		urlPrefix = strings.Trim(urlPrefix, "/")
		proxyURL = strings.TrimRight(proxyURL, "/")
		proxyURL = proxyURL + "/" + urlPrefix
	}

	return ImageReadHandler{
		fileSystemManager: fileSystemManager,
		proxyURL:          proxyURL,
	}
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

func (nfs ImageReadHandler) handleResizedImage(imagePath *filesystem.ImagePath) (http.File, error) {
	resizedFileExists, err := nfs.fileSystemManager.FileExists(imagePath, true)
	if err != nil {
		return nil, err
	}

	nonResizedFileExists, err := nfs.fileSystemManager.FileExists(imagePath, false)
	if err != nil {
		return nil, err
	}

	if !resizedFileExists {
		if nonResizedFileExists {
			return nfs.generateResizedImage(imagePath)
		}
		if nfs.proxyURL != "" {
			return nfs.getProxyImage(imagePath)
		}
	}

	return nfs.fileSystemManager.CreateFileReader(imagePath, true)
}

func (nfs ImageReadHandler) generateResizedImage(imagePath *filesystem.ImagePath) (http.File, error) {
	srcImage, err := nfs.fileSystemManager.OpenNonResizedImage(imagePath)
	if err != nil {
		return nil, err
	}

	const half = 0.5
	const MaxX = 1.0
	if imagePath.Width == 0 {
		srcBound := srcImage.Bounds()
		srcW := srcBound.Dx()
		srcH := srcBound.Dy()
		tmpW := float64(imagePath.Height) * float64(srcW) / float64(srcH)
		imagePath.Width = int(math.Max(MaxX, math.Floor(tmpW+half)))
	}

	if imagePath.Height == 0 {
		srcBound := srcImage.Bounds()
		srcW := srcBound.Dx()
		srcH := srcBound.Dy()
		tmpH := float64(imagePath.Width) * float64(srcH) / float64(srcW)
		imagePath.Height = int(math.Max(MaxX, math.Floor(tmpH+half)))
	}

	targetImg := imaging.Thumbnail(srcImage, imagePath.Width, imagePath.Height, imaging.Lanczos)

	file, err := nfs.fileSystemManager.SaveResizedImage(imagePath, targetImg)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (nfs ImageReadHandler) handleNonResizedImage(imagePath *filesystem.ImagePath) (http.File, error) {
	fileExists, err := nfs.fileSystemManager.FileExists(imagePath, false)
	if err != nil {
		return nil, err
	}

	if !fileExists {
		if nfs.proxyURL != "" {
			return nfs.getProxyImage(imagePath)
		}
		return nil, nfs.createNonExistsError(imagePath.ImageFile)
	}

	return nfs.fileSystemManager.CreateFileReader(imagePath, false)
}

func (nfs ImageReadHandler) getProxyImage(imagePath *filesystem.ImagePath) (http.File, error) {
	if imagePath.RawResizedFolder == "" {
		return DownloadFile(
			fmt.Sprintf("%s/%s/%s", nfs.proxyURL, imagePath.FolderName, imagePath.ImageFile),
			imagePath.ImageFile,
		)
	}

	return DownloadFile(
		fmt.Sprintf(
			"%s/%s/%s/%s",
			nfs.proxyURL,
			imagePath.RawResizedFolder,
			imagePath.FolderName,
			imagePath.ImageFile,
		),
		imagePath.ImageFile,
	)
}
