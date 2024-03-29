package assets

import (
	"fmt"
	"net/http"

	"github.com/breathbath/media-library/authentication"

	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-library/filesystem"
	"github.com/gorilla/mux"
)

type ImageDeleteHandler struct {
	FileSystemManager filesystem.LocalFileSystemManager
}

func (idh ImageDeleteHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) { //nolint:gocyclo
	token := r.Context().Value(authentication.TokenContextKey)
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	folder, ok := vars["folder"]

	if !ok {
		io.OutputWarning("", "'folder' parameter not provide in uri")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	imageName, ok := vars["image"]

	if !ok {
		io.OutputWarning("", "'image' parameter not provide in uri")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	proxyURL := env.ReadEnv("PROXY_URL", "")
	imagePathRaw := folder + "/" + imageName
	imagePath := parseImagePath(imagePathRaw)
	if !imagePath.IsValid {
		io.OutputError(fmt.Errorf("failed to parse image url %s", imagePathRaw), "", "")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// removing non resized image e.g. /images/ldjfksljfas/someImage.png
	nonResizedImageDeletionErr := idh.FileSystemManager.RemoveNonResizedImage(imagePath)
	statusToGive := http.StatusOK
	if nonResizedImageDeletionErr != nil {
		if idh.FileSystemManager.IsNonExistingPathError(nonResizedImageDeletionErr) {
			if proxyURL == "" {
				statusToGive = http.StatusNotFound
			} else {
				io.OutputWarning(
					"",
					"Cannot find resized image to delete, since proxy param is set, "+
						"this image might be available on proxy server, therefore error is ignored",
				)
			}
		} else {
			io.OutputError(nonResizedImageDeletionErr, "", "Failed to delete non-resized file '%s'", imagePath.ImageFile)
			statusToGive = http.StatusInternalServerError
		}
	}

	// removing resized folder e.g. /images/cache/resized_image/ldjfksljfas/someImage
	resizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, true, false)
	if resizedFolderDeletionErr != nil && !idh.FileSystemManager.IsNonExistingPathError(resizedFolderDeletionErr) {
		io.OutputError(resizedFolderDeletionErr, "", "Failed to delete resized folder for file '%s'", imagePath.ImageFile)
		if statusToGive == http.StatusOK {
			statusToGive = http.StatusInternalServerError
		}
	}

	// check if /images/ldjfksljfas is empty (it could contain other images)
	isDirEmpty, dirListingErr := idh.FileSystemManager.IsImageDirEmpty(imagePath, false)
	// removing non resized folder e.g. /images/ldjfksljfas if it's empty (it could contain other images)
	if dirListingErr != nil {
		if proxyURL == "" {
			if statusToGive == http.StatusOK {
				statusToGive = http.StatusInternalServerError
				io.OutputError(dirListingErr, "", "Failed to list directory '%s'", imagePath.FolderName)
			}
		}
	} else if isDirEmpty {
		nonResizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, false, false)
		if nonResizedFolderDeletionErr != nil && proxyURL == "" {
			io.OutputError(nonResizedFolderDeletionErr, "", "Failed to delete non-resized folder for file '%s'", imagePath.ImageFile)
			if statusToGive == http.StatusOK {
				statusToGive = http.StatusInternalServerError
			}
		}
	}

	// check if /images/cache/resized_image/ldjfksljfas is empty (it could contain other folders)
	isDirEmpty, dirListingErr = idh.FileSystemManager.IsImageDirEmpty(imagePath, true)
	if dirListingErr != nil {
		if proxyURL == "" && statusToGive == http.StatusOK {
			statusToGive = http.StatusInternalServerError
			io.OutputError(dirListingErr, "", "Failed to list resized directory '%s'", imagePath.FolderName)
		}
	} else if isDirEmpty {
		nonResizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, true, true)
		if nonResizedFolderDeletionErr != nil {
			if proxyURL == "" && statusToGive == http.StatusOK {
				io.OutputError(nonResizedFolderDeletionErr, "", "Failed to delete resized parent folder for file '%s'", imagePath.ImageFile)
				statusToGive = http.StatusInternalServerError
			}
		}
	}

	rw.WriteHeader(statusToGive)
}
