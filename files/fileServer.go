package files

import (
	"fmt"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/fs"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	io2 "io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const UNIQUE_FOLDER_ATTEMPTS_COUNT = 5

type FileSystemManager struct {
	assetsPath string
}

func NewFileSystemManager(assetsPath string) FileSystemManager {
	return FileSystemManager{assetsPath: assetsPath}
}

func (nfs FileSystemManager) HandlePost(rw http.ResponseWriter, r *http.Request) {
	var err error

	token := r.Context().Value("token")
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	maxUploadFileSizeMb := env.ReadEnvInt("MAX_UPLOADED_FILE_MB", 7)
	err = r.ParseMultipartForm(maxUploadFileSizeMb << 20)
	if err != nil {
		io.OutputError(err, "", "Multipart form parse failure")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadedFiles, ok := r.MultipartForm.File["files"]
	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	imgPath := ImagePath{isValid: false}
	for _, uploadedFileHeader := range uploadedFiles {
		infile, err := uploadedFileHeader.Open()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded source file opening failure")
			return
		}

		fileName := uploadedFileHeader.Filename

		if filepath.Ext(fileName) == "" {
			_, ext, err := mimetype.DetectReader(infile)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				io.OutputError(err, "", "Failed to detect mimetype and extension of uploaded file '%s'", fileName)
				return
			}
			fileName += "." + ext
		}

		if !imgPath.isValid {
			imgPath, err = nfs.generateUniqueImagePath(fileName)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				io.OutputError(err, "", "Folder generation failure")
				return
			}
		} else {
			imgPath = parseImagePath(imgPath.folderName+"/"+fileName, nfs.assetsPath)
		}

		if !imgPath.isValid {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		outfile, err := os.Create(imgPath.GetNonResizedImagePath())
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded target file opening failure")
			return
		}

		_, err = io2.Copy(outfile, infile)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded file copy failure")
			return
		}
	}
}

func (nfs FileSystemManager) GetFileContentType(out *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func (nfs FileSystemManager) uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000
	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

func (nfs FileSystemManager) generateUniqueImagePath(fileName string) (ImagePath, error) {
	var imgPath ImagePath
	for i := 0; i < UNIQUE_FOLDER_ATTEMPTS_COUNT; i++ {
		uniqueFolderName := nfs.uniqid("")
		imgPath = parseImagePath(filepath.Join(uniqueFolderName, fileName), nfs.assetsPath)
		nonResizedFolderPath := imgPath.GetNonResizedFolderPath()
		if !fs.FileExists(nonResizedFolderPath) {
			return imgPath, nil
		}
	}
	nonResizedImagePath := imgPath.GetNonResizedImagePath()
	if !fs.FileExists(nonResizedImagePath) {
		return imgPath, nil
	}

	return imgPath, fmt.Errorf("Cannot generate unique folder at '%s'", imgPath.GetNonResizedFolderPath())
}

func (nfs FileSystemManager) HandleDelete(rw http.ResponseWriter, r *http.Request) {
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

	imagePath := parseImagePath(folder+"/"+imageName, nfs.assetsPath)
	if !imagePath.isValid {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	nonResizedFilePath := imagePath.GetNonResizedImagePath()
	nonResizedDeletionErr := os.Remove(nonResizedFilePath)

	resizedFolderPath := imagePath.GetResizedFolderPath()
	resizedFolderDeletionErr := os.RemoveAll(resizedFolderPath)

	status := http.StatusOK

	if nonResizedDeletionErr != nil {
		if os.IsNotExist(nonResizedDeletionErr) {
			status = http.StatusNotFound
		} else {
			io.OutputError(nonResizedDeletionErr, "", "Failed to delete file '%s'", nonResizedFilePath)
		}
	} else {
		nonResizedFolderPath := imagePath.GetNonResizedFolderPath()
		isDirEmpty, dirListingErr := nfs.isDirEmpty(nonResizedFolderPath)
		if dirListingErr != nil {
			io.OutputError(dirListingErr, "", "Failed to list directory '%s'", nonResizedFolderPath)
		} else if isDirEmpty {
			nonResizedFolderDeletionErr := os.RemoveAll(nonResizedFolderPath)
			if nonResizedFolderDeletionErr != nil {
				io.OutputError(nonResizedDeletionErr, "", "Failed to delete directory '%s'", nonResizedDeletionErr)
			}
		}
	}

	if resizedFolderDeletionErr != nil {
		io.OutputError(resizedFolderDeletionErr, "", "Failed to delete folder '%s'", resizedFolderPath)
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}
	}

	rw.WriteHeader(status)
}

func (nfs FileSystemManager) Open(path string) (http.File, error) {
	imagePath := parseImagePath(path, nfs.assetsPath)
	if !imagePath.isValid {
		return nil, nfs.createNonExistsError(path)
	}

	if imagePath.rawResizedFolder != "" {
		return nfs.handleResizedImage(imagePath)
	}

	return nfs.handleNonResizedImage(imagePath)
}

func (nfs FileSystemManager) createNonExistsError(path string) *os.PathError {
	return &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (nfs FileSystemManager) handleResizedImage(imagePath ImagePath) (http.File, error) {
	resizedPath := imagePath.GetResizedImagePath()
	info, err := os.Stat(resizedPath)
	if os.IsNotExist(err) {
		f, err := nfs.generateResizedImage(imagePath)
		return f, err
	}

	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nfs.createNonExistsError(resizedPath)
	}

	f, err := os.Open(resizedPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (nfs FileSystemManager) isDirEmpty(name string) (bool, error) {
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

func (nfs FileSystemManager) generateResizedImage(imagePath ImagePath) (http.File, error) {
	resizedFolder := imagePath.GetResizedFolderPath()
	resizedPath := imagePath.GetResizedImagePath()

	err := os.MkdirAll(resizedFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	nonResizedImagePath := imagePath.GetNonResizedImagePath()
	_, err = os.Stat(nonResizedImagePath)

	if err != nil {
		return nil, err
	}

	src, err := imaging.Open(nonResizedImagePath)
	if err != nil {
		return nil, err
	}

	src = imaging.Fill(src, imagePath.width, imagePath.height, imaging.Center, imaging.Lanczos)

	err = imaging.Save(src, resizedPath)
	if err != nil {
		return nil, err
	}

	return os.Open(resizedPath)
}

func (nfs FileSystemManager) handleNonResizedImage(imagePath ImagePath) (http.File, error) {
	fullImagePath := imagePath.GetNonResizedImagePath()

	info, err := os.Stat(fullImagePath)
	if os.IsNotExist(err) {
		return nil, nfs.createNonExistsError(fullImagePath)
	}
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nfs.createNonExistsError(fullImagePath)
	}

	f, err := os.Open(fullImagePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}
