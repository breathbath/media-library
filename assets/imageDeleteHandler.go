package assets

import (
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-library/fileSystem"
	"github.com/gorilla/mux"
	"net/http"
)

type ImageDeleteHandler struct {
	FileSystemManager fileSystem.LocalFileSystemManager
}

func (idh ImageDeleteHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	token := r.Context().Value("token")
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	folder, ok := vars["folder"]

	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	imageName, ok := vars["image"]

	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	imagePath := parseImagePath(folder + "/" + imageName)
	if !imagePath.IsValid {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	status := http.StatusOK
	//removing non resized image e.g. /images/ldjfksljfas/someImage.png
	nonResizedImageDeletionErr := idh.FileSystemManager.RemoveNonResizedImage(imagePath)
	if nonResizedImageDeletionErr != nil {
		if idh.FileSystemManager.IsNonExistingPathError(nonResizedImageDeletionErr) {
			status = http.StatusNotFound
		} else {
			io.OutputError(nonResizedImageDeletionErr, "", "Failed to delete non-resized file '%s'", imagePath.ImageFile)
		}
	}

	//removing resized folder e.g. /images/cache/resized_image/ldjfksljfas/someImage
	resizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, true, false)
	if resizedFolderDeletionErr != nil {
		io.OutputError(resizedFolderDeletionErr, "", "Failed to delete resized folder for file '%s'", imagePath.ImageFile)
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}
	}

	//check if /images/ldjfksljfas is empty (it could contain other images)
	isDirEmpty, dirListingErr := idh.FileSystemManager.IsImageDirEmpty(imagePath, false)
	//removing non resized folder e.g. /images/ldjfksljfas if it's empty (it could contain other images)
	if dirListingErr != nil {
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}
		io.OutputError(dirListingErr, "", "Failed to list directory '%s'", imagePath.FolderName)
	} else if isDirEmpty {
		nonResizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, false, false)
		if nonResizedFolderDeletionErr != nil {
			io.OutputError(nonResizedFolderDeletionErr, "", "Failed to delete non-resized folder for file '%s'", imagePath.ImageFile)
			if status == http.StatusOK {
				status = http.StatusInternalServerError
			}
		}
	}

	//check if /images/cache/resized_image/ldjfksljfas is empty (it could contain other folders)
	isDirEmpty, dirListingErr = idh.FileSystemManager.IsImageDirEmpty(imagePath, true)
	if dirListingErr != nil {
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}
		io.OutputError(dirListingErr, "", "Failed to list resized directory '%s'", imagePath.FolderName)
	} else if isDirEmpty {
		nonResizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, true, true)
		if nonResizedFolderDeletionErr != nil {
			io.OutputError(nonResizedFolderDeletionErr, "", "Failed to delete resized parent folder for file '%s'", imagePath.ImageFile)
			if status == http.StatusOK {
				status = http.StatusInternalServerError
			}
		}
	}

	rw.WriteHeader(status)
}
