package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/go_utils/utils/fs"
	"github.com/breathbath/media-service/test/helper"
	"github.com/stretchr/testify/assert"
	"image"
	"io"
	"io/ioutil"
	http2 "net/http"
	"os"
	"path/filepath"
	"testing"
)

type filesResponse struct {
	FilesToReturn []string `json:"filepathes"`
}

func TestPostHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	setup()
	defer cleanup()
	t.Run("testImageSaved", testImageSaved)
	t.Run("testInvalidTokenForPosting", testInvalidTokenForPosting)
	t.Run("testTooLargeFile", testTooLargeFile)
	t.Run("testWrongMime", testWrongMime)
	t.Run("testTooBigDimension", testTooBigDimension)
	t.Run("testDelete", testDelete)
	t.Run("testInvalidTokenForDeleting", testInvalidTokenForDeleting)
}

func testImageSaved(t *testing.T) {
	jpgImage, err := helper.CreateImage(helper.ImageSpec{Format: "jpg"})
	assert.NoError(t, err)

	pngImage, err := helper.CreateImage(helper.ImageSpec{Format: "png"})
	assert.NoError(t, err)

	var filesResp filesResponse
	statusCode, err := makeTestingPost(
		&filesResp,
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "some Image.jpg",
			File:      jpgImage,
		},
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "someImage.png",
			File:      pngImage,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)
	assert.Len(t, filesResp.FilesToReturn, 2)

	assert.Regexp(t, `[\w]+/some_Image.jpg`, filesResp.FilesToReturn[0])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[0]))

	assert.Regexp(t, `[\w]+/someImage.png`, filesResp.FilesToReturn[1])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[1]))

	pngImage2, err := helper.CreateImage(helper.ImageSpec{Format: "png"})
	assert.NoError(t, err)

	var filesResp2 filesResponse
	statusCode2, err := makeTestingPost(
		&filesResp2,
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "someImage.png",
			File:      pngImage2,
		},
	)
	assert.Equal(t, http2.StatusOK, statusCode2)
	assert.NoError(t, err)
	assert.Len(t, filesResp2.FilesToReturn, 1)
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp2.FilesToReturn[0]))

	//asserting that each upload saves files with same names to different locations
	assert.NotEqual(t, filesResp.FilesToReturn[1], filesResp2.FilesToReturn[0])
}

func testTooLargeFile(t *testing.T) {
	img, err := helper.CreateImage(helper.ImageSpec{Format: "jpg", Width: 2000, Height: 2000, Quality: 100})
	assert.NoError(t, err)

	bodyBytes, err := ioutil.ReadAll(img)
	assert.NoError(t, err)

	var filesError struct{ Files []string `json:"files"` }
	statusCode, err := makeTestingPost(
		&filesError,
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "someImage.png",
			File:      ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusBadRequest, statusCode)
	assert.Len(t, filesError.Files, 1)

	expectedOutput := fmt.Sprintf(
		"The file is too large (%d bytes). Allowed maximum size is %v Mb",
		len(bodyBytes),
		0.1,
	)
	assert.Equal(t, expectedOutput, filesError.Files[0])
}

func testWrongMime(t *testing.T) {
	textReader := bytes.NewBuffer([]byte("some text"))

	var filesError struct{ Files []string `json:"files"` }
	statusCode, err := makeTestingPost(
		&filesError,
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "someImage.png",
			File:      textReader,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusBadRequest, statusCode)
	assert.Len(t, filesError.Files, 1)

	assert.Equal(t, "Not supported image type 'text/plain', supported types are jpg|jpeg|png|gif", filesError.Files[0])
}

func testTooBigDimension(t *testing.T) {
	img, err := helper.CreateImage(helper.ImageSpec{Format: "jpg", Width: 1000, Height: 600})
	assert.NoError(t, err)

	var filesResp filesResponse
	statusCode, err := makeTestingPost(
		&filesResp,
		helper.UploadedFile{
			FieldName: "files",
			FileName:  "someImage.jpg",
			File:      img,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)

	filePath := filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[0])

	file, err := os.Open(filePath)
	assert.NoError(t, err)

	imgConfig, _, err := image.DecodeConfig(file)
	assert.NoError(t, err)
	assert.Equal(t, 833, int(imgConfig.Width))
	assert.Equal(t, 500, int(imgConfig.Height))
}

func testInvalidTokenForPosting(t *testing.T) {
	testClient := helper.NewTestClient()
	statusCode, _, err := testClient.MakePost("", "http://localhost:9925/images")
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusForbidden, statusCode)
}

func testInvalidTokenForDeleting(t *testing.T) {
	testClient := helper.NewTestClient()
	statusCode, err := testClient.MakeDelete("", "http://localhost:9925/images/lala/mama")
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusForbidden, statusCode)
}

func testDelete(t *testing.T) {
	err := saveImage("lsls", "someImg.png")
	assert.NoError(t, err)

	err = saveImage(filepath.Join("cache", "resized_image", "lsls", "someImg"), "5x5.png")
	assert.NoError(t, err)

	testClient := helper.NewTestClient()
	validToken, err := testClient.GenerateValidToken()
	if err != nil {
		return
	}

	statusCode, err := testClient.MakeDelete(validToken, "http://localhost:9925/images/lsls/someImg.png")

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)
	assert.False(t, fs.FileExists(filepath.Join(helper.AssetsPath, "lsls")))
	assert.False(t, fs.FileExists(filepath.Join(helper.AssetsPath, "lsls", "someImg.png")))
	assert.False(t, fs.FileExists(filepath.Join(helper.AssetsPath, "cache", "resized_image", "lsls", "someImg", "5x5.png")))
	assert.False(t, fs.FileExists(filepath.Join(helper.AssetsPath, "cache", "resized_image", "lsls", "someImg")))
	assert.False(t, fs.FileExists(filepath.Join(helper.AssetsPath, "cache", "resized_image", "lsls")))
}

func saveImage(folderPath, imageName string) error {
	img, err := helper.CreateImage(helper.ImageSpec{Format: "png", Width: 10, Height: 10})
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(helper.AssetsPath, folderPath), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(helper.AssetsPath, folderPath, imageName))
	if err != nil {
		return err
	}

	_, err = io.Copy(f, img)
	if err != nil {
		return err
	}

	return nil
}

func makeTestingPost(target interface{}, files ...helper.UploadedFile) (statusCode int, err error) {
	testClient := helper.NewTestClient()
	err = testClient.AddFiles(files...)
	if err != nil {
		return
	}

	validToken, err := testClient.GenerateValidToken()
	if err != nil {
		return
	}

	statusCode, body, err := testClient.MakePost(validToken, "http://localhost:9925/images")
	if err != nil {
		return
	}
	fmt.Println(body)

	err = json.Unmarshal([]byte(body), target)
	if err != nil {
		return
	}

	return
}

func setup() {
	err := helper.PrepareFileServer()
	errs.FailOnError(err)
}

func cleanup() {
	helper.ShutdownFileServer()
}
