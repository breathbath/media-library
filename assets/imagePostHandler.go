package assets

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/breathbath/media-library/authentication"

	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	error2 "github.com/breathbath/media-library/error"
	"github.com/gabriel-vasile/mimetype"
)

const SubmittedFileFieldName = "files[]"

type ImagePostHandler struct {
	ImageSaver          ImageSaver
	maxUploadFileSizeMb float64
}

func uniqid() string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000
	return fmt.Sprintf("%08x%05x", sec, usec)
}

func NewImagePostHandler(imgSaver ImageSaver) ImagePostHandler {
	return ImagePostHandler{
		ImageSaver:          imgSaver,
		maxUploadFileSizeMb: env.ReadEnvFloat("MAX_UPLOADED_FILE_MB", 20),
	}
}

func (iph ImagePostHandler) HandlePost(rw http.ResponseWriter, r *http.Request) { // nolint:funlen
	rw.Header().Set("Content-Type", "application/json")

	token := r.Context().Value(authentication.TokenContextKey)
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	const twentyMb = 20
	err := r.ParseMultipartForm(int64(iph.maxUploadFileSizeMb) * 3 << twentyMb)
	if err != nil {
		io.OutputError(err, "", "Multipart form parse failure")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadedFiles, ok := r.MultipartForm.File[SubmittedFileFieldName]
	if !ok {
		submittedNames := ""
		for submittedName := range r.MultipartForm.File {
			submittedNames += submittedName + ", "
		}

		io.OutputWarning(
			"",
			"MultipartForm file field '%s' is not submitted, submitted fields list %s",
			SubmittedFileFieldName,
			submittedNames,
		)

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	filesToReturn := make([]string, 0, len(uploadedFiles))
	validationErrors := error2.NewValidationErrors()
	folderName := uniqid()
	for _, uploadedFileHeader := range uploadedFiles {
		io.OutputInfo(
			"",
			"Got file to save: name: %s, size: %d bytes, header: %v",
			uploadedFileHeader.Filename,
			uploadedFileHeader.Size,
			uploadedFileHeader.Header,
		)
		statusErr, curFilesToReturn := iph.handleUploadedFile(uploadedFileHeader, iph.maxUploadFileSizeMb, folderName)
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
		var body []byte
		body, err = json.Marshal(validationErrors)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "cannot generate response for validation errors")
			return
		}
		io.OutputError(fmt.Errorf("validation errors for incoming file: %s", string(body)), "", "Validation failure")
		rw.WriteHeader(http.StatusBadRequest)
		_, err = rw.Write(body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "cannot send json data")
		}

		return
	}

	if len(filesToReturn) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		io.OutputWarning("", "No files were submitted")
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
) (statusErr error2.StatusError, files []string) {
	filesToReturn := []string{}
	infile, err := uploadedFileHeader.Open()
	defer func() {
		err = infile.Close()
		if err != nil {
			io.OutputError(err, "", "Failed to close input file")
		}
	}()

	if err != nil {
		return error2.StatusError{
			Status: http.StatusInternalServerError,
			Error:  err,
			Text:   "Uploaded source file opening failure",
		}, filesToReturn
	}

	fileName := uploadedFileHeader.Filename
	detectedMime, detectedExt, err := mimetype.DetectReader(infile)
	if err != nil {
		return error2.StatusError{
			Status: http.StatusBadRequest,
			Error:  err,
			Text:   fmt.Sprintf("Failed to detect mimetype and extension of uploaded file '%s'", fileName),
		}, filesToReturn
	}

	validationErrs, err := Validate(uploadedFileHeader, detectedMime, SubmittedFileFieldName, maxUploadFileSizeMb)
	if err != nil {
		return error2.StatusError{
			Status: http.StatusBadRequest,
			Error:  err,
			Text:   fmt.Sprintf("Failed to detect mimetype and extension of uploaded file '%s'", fileName),
		}, filesToReturn
	}

	if len(validationErrs) > 0 {
		return error2.StatusError{
			Status:         http.StatusBadRequest,
			Error:          nil,
			ValidationErrs: validationErrs,
		}, filesToReturn
	}

	if filepath.Ext(fileName) == "" {
		if detectedExt != "" {
			fileName += "." + detectedExt
			io.OutputInfo("", "Added extension to the file: %s", fileName)
		} else {
			io.OutputWarning("", "Was not able to detect file extension")
		}
	}

	fileName = SanitizeImageName(fileName)
	io.OutputInfo("", "File name after sanitizing: %s", fileName)

	err = iph.ImageSaver.SaveImage(infile, folderName, fileName)
	if err != nil {
		return error2.StatusError{
			Status:         http.StatusInternalServerError,
			Error:          err,
			ValidationErrs: validationErrs,
			Text:           "Folder generation failure",
		}, filesToReturn
	}

	filesToReturn = append(filesToReturn, folderName+"/"+fileName)

	return error2.StatusError{}, filesToReturn
}
