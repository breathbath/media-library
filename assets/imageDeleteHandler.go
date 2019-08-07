package assets

import (
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-service/fileSystem"
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

	imagePath := parseImagePath(folder+"/"+imageName)
	if !imagePath.IsValid {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	nonResizedDeletionErr := idh.FileSystemManager.RemoveNonResizedImage(imagePath)
	resizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, true)

	status := http.StatusOK

	if nonResizedDeletionErr != nil {
		if idh.FileSystemManager.IsNonExistingPathError(nonResizedDeletionErr) {
			status = http.StatusNotFound
		} else {
			io.OutputError(nonResizedDeletionErr, "", "Failed to delete non resized file '%s'", imagePath.ImageFile)
		}
	} else {
		isDirEmpty, dirListingErr := idh.FileSystemManager.IsNonResizedImageDirEmpty(imagePath)
		if dirListingErr != nil {
			io.OutputError(dirListingErr, "", "Failed to list directory '%s'", imagePath.FolderName)
		} else if isDirEmpty {
			nonResizedFolderDeletionErr := idh.FileSystemManager.RemoveDir(imagePath, false)
			if nonResizedFolderDeletionErr != nil {
				io.OutputError(nonResizedDeletionErr, "", "Failed to delete directory '%s'", imagePath.FolderName)
			}
		}
	}

	if resizedFolderDeletionErr != nil {
		io.OutputError(resizedFolderDeletionErr, "", "Failed to delete resized images folder '%s'", imagePath.FolderName)
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}
	}

	rw.WriteHeader(status)
}