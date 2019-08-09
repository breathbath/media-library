package assets

import (
	"encoding/json"
	"fmt"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	error2 "github.com/breathbath/media-service/error"
	"github.com/gabriel-vasile/mimetype"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

const SubmittedFileFieldName = "files"

type ImagePostHandler struct {
	ImageSaver ImageSaver
}

func uniqid() string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000
	return fmt.Sprintf("%08x%05x", sec, usec)
}

func (iph ImagePostHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {
	var err error

	token := r.Context().Value("token")
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	maxUploadFileSizeMb := env.ReadEnvFloat("MAX_UPLOADED_FILE_MB", 7)
	err = r.ParseMultipartForm(int64(maxUploadFileSizeMb) * 3 << 20)
	if err != nil {
		io.OutputError(err, "", "Multipart form parse failure")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadedFiles, ok := r.MultipartForm.File[SubmittedFileFieldName]
	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	filesToReturn := make([]string, 0, len(uploadedFiles))
	validationErrors := error2.NewValidationErrors()
	folderName := uniqid()
	for _, uploadedFileHeader := range uploadedFiles {
		statusErr, curFilesToReturn := iph.handleUploadedFile(uploadedFileHeader, maxUploadFileSizeMb, folderName)
		if statusErr.Error != nil {
			rw.WriteHeader(statusErr.Status)
			io.OutputError(statusErr.Error, "", statusErr.Text)
			return
		}

		if len(statusErr.ValidationErrs) > 0 {
			validationErrors.Merge(statusErr.ValidationErrs)
		}

		filesToReturn = append(filesToReturn, curFilesToReturn...)
	}

	if len(validationErrors) > 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(rw).Encode(validationErrors)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Cannot send json data")
		}
		return
	}

	if len(filesToReturn) == 0 {
		rw.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(rw).Encode(map[string][]string{
			SubmittedFileFieldName: {"Should contain at least 1 element"},
		})
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Cannot send json data")
		}
		return
	}

	resp := struct {
		FilesToReturn []string `json:"filepathes"`
	}{
		FilesToReturn: filesToReturn,
	}

	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		io.OutputError(err, "", "Cannot send json data")
		return
	}
}

func (iph ImagePostHandler) handleUploadedFile(
	uploadedFileHeader *multipart.FileHeader,
	maxUploadFileSizeMb float64,
	folderName string,
) (error2.StatusError, []string) {
	filesToReturn := []string{}
	infile, err := uploadedFileHeader.Open()
	defer infile.Close()

	if err != nil {
		return error2.StatusError{
			Status: http.StatusInternalServerError,
			Error:  err,
			Text:   "Uploaded source file opening failure",
		}, filesToReturn
	}

	fileName := uploadedFileHeader.Filename

	validationErrs, err := Validate(uploadedFileHeader, infile, SubmittedFileFieldName, maxUploadFileSizeMb)
	if err != nil {
		return error2.StatusError{
			Status: http.StatusBadRequest,
			Error:  err,
			Text:   fmt.Sprintf("Failed to detect mimetype and extension of uploaded file '%s'", fileName),
		}, filesToReturn
	}

	if len(validationErrs) > 0 {
		return error2.StatusError{
			Status: http.StatusBadRequest,
			Error:  nil,
			ValidationErrs: validationErrs,
		}, filesToReturn
	}

	if filepath.Ext(fileName) == "" {
		_, ext, err := mimetype.DetectReader(infile)
		if err != nil {
			return error2.StatusError{
				Status: http.StatusBadRequest,
				Error:  err,
				ValidationErrs: validationErrs,
				Text: fmt.Sprintf("Failed to detect mimetype and extension of uploaded file '%s'", fileName),
			}, filesToReturn
		}
		fileName += "." + ext
	}

	fileName = SanitizeImageName(fileName)

	err = iph.ImageSaver.SaveImage(infile, folderName, fileName)
	if err != nil {
		return error2.StatusError{
			Status: http.StatusInternalServerError,
			Error:  err,
			ValidationErrs: validationErrs,
			Text: "Folder generation failure",
		}, filesToReturn
	}

	filesToReturn = append(filesToReturn, folderName+"/"+fileName)

	return error2.StatusError{}, filesToReturn
}
