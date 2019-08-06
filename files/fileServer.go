package files

import (
	"encoding/json"
	"fmt"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/fs"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	io2 "io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type FileSystemManager struct {
	assetsPath string
}

func NewFileSystemManager(assetsPath string) FileSystemManager {
	return FileSystemManager{assetsPath: assetsPath}
}

func (nfs FileSystemManager) saveCompressedImageIfPossible(sourceFile multipart.File, targetFile io2.Writer, ext string) error {
	_, err := sourceFile.Seek(0, 0)
	if err != nil {
		return err
	}

	imgRcr, _, err := image.Decode(sourceFile)
	if err != nil {
		return err
	}

	resizeX, resizeY := 0, 0
	bounds := imgRcr.Bounds()
	if (bounds.Dy() > bounds.Dx() || bounds.Dy() == bounds.Dx()) && bounds.Dx() > VERT_MAX_IMAGE_WIDTH {
		resizeX = VERT_MAX_IMAGE_WIDTH
	} else if bounds.Dx() > bounds.Dy() && bounds.Dy() > HORIZ_MAX_IMAGE_HEIGHT {
		resizeY = HORIZ_MAX_IMAGE_HEIGHT
	}

	if resizeX + resizeY > 0 {
		imgRcr = imaging.Resize(imgRcr, resizeX, resizeY, imaging.Lanczos)
	}

	if ext == "jpg" || ext == "jpeg" {
		jpegQuality := env.ReadEnvInt("COMPRESS_JPG_QUALITY", 85)
		return jpeg.Encode(targetFile, imgRcr, &jpeg.Options{int(jpegQuality)})
	}

	if ext == "png" {
		encoder := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}
		return encoder.Encode(targetFile, imgRcr)
	}

	if ext == "gif" {
		return gif.Encode(targetFile, imgRcr, &gif.Options{})
	}

	_, err = io2.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func (nfs FileSystemManager) HandlePost(rw http.ResponseWriter, r *http.Request) {
	var err error

	token := r.Context().Value("token")
	if token == nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	maxUploadFileSizeMb := env.ReadEnvInt("MAX_UPLOADED_FILE_MB", 7)
	err = r.ParseMultipartForm(maxUploadFileSizeMb * 3 << 20)
	if err != nil {
		io.OutputError(err, "", "Multipart form parse failure")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	submittedFileFieldName := "files"
	uploadedFiles, ok := r.MultipartForm.File[submittedFileFieldName]
	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	imgPath := ImagePath{isValid: false}
	filesToReturn := make([]string, 0, len(uploadedFiles))
	for _, uploadedFileHeader := range uploadedFiles {
		infile, err := uploadedFileHeader.Open()
		defer infile.Close()

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded source file opening failure")
			return
		}

		fileName := uploadedFileHeader.Filename

		validationErrs, err := Validate(uploadedFileHeader, infile, submittedFileFieldName, maxUploadFileSizeMb)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			io.OutputError(err, "", "Failed to detect mimetype and extension of uploaded file '%s'", fileName)
			return
		}

		if len(validationErrs) > 0 {
			rw.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(rw).Encode(validationErrs)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				io.OutputError(err, "", "Cannot send json data")
			}
			return
		}

		if filepath.Ext(fileName) == "" {
			_, ext, err := mimetype.DetectReader(infile)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				io.OutputError(err, "", "Failed to detect mimetype and extension of uploaded file '%s'", fileName)
				return
			}
			fileName += "." + ext
		}

		fileName = nfs.sanitizeImageName(fileName)

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

		targetDir := imgPath.GetNonResizedFolderPath()
		err = os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Failed to create directory '%s' for uploaded files", targetDir)
			return
		}

		outfile, err := os.Create(imgPath.GetNonResizedImagePath())
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded target file opening failure")
			return
		}
		defer outfile.Close()

		err = nfs.saveCompressedImageIfPossible(infile, outfile, imgPath.imageExt)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			io.OutputError(err, "", "Uploaded file copy failure")
			return
		}

		filesToReturn = append(filesToReturn, imgPath.folderName+"/"+imgPath.imageFile)
	}

	if len(filesToReturn) == 0 {
		rw.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(rw).Encode(map[string][]string{
			submittedFileFieldName: {"Should contain at least 1 element"},
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

func (nfs FileSystemManager) sanitizeImageName(fullName string) string {
	ext := filepath.Ext(fullName)
	imageName := fullName[0 : len(fullName)-len(ext)]
	r := regexp.MustCompile(`[^\w\-_]`)
	sanitizedImageName := r.ReplaceAllString(imageName, "_")

	return sanitizedImageName + ext
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
