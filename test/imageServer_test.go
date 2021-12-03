package test

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	http2 "net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/go_utils/utils/fs"
	"github.com/breathbath/media-library/test/helper"
	"github.com/stretchr/testify/assert"
)

type filesResponse struct {
	FilesToReturn []string `json:"filepathes"`
}

func TestHandlers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	setup()
	defer cleanup()
	t.Run("testImageSaved", testImageSaved)
	t.Run("testInvalidTokenForPosting", testInvalidTokenForPosting)
	t.Run("testTooLargeFile", testTooLargeFile)
	t.Run("testWrongMime", testWrongMime)
	t.Run("testNoExtension", testNoExtension)
	t.Run("testTooBigDimension", testTooBigDimension)
	t.Run("testDelete", testDelete)
	t.Run("testInvalidTokenForDeleting", testInvalidTokenForDeleting)
	t.Run("testRead", testRead)
	t.Run("testReadNonExistingImage", testReadNonExistingImage)
	t.Run("testGettingResizedImage", testGettingResizedImage)
	t.Run("testGettingCachedResizedImage", testGettingCachedResizedImage)

	t.Run("testProxyMatch", testProxyMatch)
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
			FieldName: "files[]",
			FileName:  "some Image.jpg",
			File:      jpgImage,
		},
		helper.UploadedFile{
			FieldName: "files[]",
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
			FieldName: "files[]",
			FileName:  "someImage.png",
			File:      pngImage2,
		},
	)
	assert.Equal(t, http2.StatusOK, statusCode2)
	assert.NoError(t, err)
	assert.Len(t, filesResp2.FilesToReturn, 1)
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp2.FilesToReturn[0]))

	// asserting that each upload saves files with same names to different locations
	assert.NotEqual(t, filesResp.FilesToReturn[1], filesResp2.FilesToReturn[0])
}

func testTooLargeFile(t *testing.T) {
	img, err := helper.CreateImage(helper.ImageSpec{Format: "jpg", Width: 2000, Height: 2000, Quality: 100})
	assert.NoError(t, err)

	bodyBytes, err := ioutil.ReadAll(img)
	assert.NoError(t, err)

	var filesError struct {
		Files []string `json:"files[]"`
	}
	statusCode, err := makeTestingPost(
		&filesError,
		helper.UploadedFile{
			FieldName: "files[]",
			FileName:  "someImage.png",
			File:      ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusBadRequest, statusCode)
	assert.Len(t, filesError.Files, 1)

	if len(filesError.Files) > 0 {
		expectedOutput := fmt.Sprintf(
			"The file is too large (%d bytes). Allowed maximum size is %v Mb",
			len(bodyBytes),
			0.1,
		)
		assert.Equal(t, expectedOutput, filesError.Files[0])
	}
}

func testWrongMime(t *testing.T) {
	textReader := bytes.NewBuffer([]byte("some text"))

	var filesError struct {
		Files []string `json:"files[]"`
	}
	statusCode, err := makeTestingPost(
		&filesError,
		helper.UploadedFile{
			FieldName: "files[]",
			FileName:  "someImage.png",
			File:      textReader,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusBadRequest, statusCode)
	assert.Len(t, filesError.Files, 1)

	if len(filesError.Files) > 0 {
		assert.Equal(t, "Not supported image type 'text/plain', supported types are jpg|jpeg|png|gif", filesError.Files[0])
	}
}

func testNoExtension(t *testing.T) {
	jpgImage, err := helper.CreateImage(helper.ImageSpec{Format: "jpg"})
	assert.NoError(t, err)

	pngImage, err := helper.CreateImage(helper.ImageSpec{Format: "png"})
	assert.NoError(t, err)

	var filesResp filesResponse
	statusCode, err := makeTestingPost(
		&filesResp,
		helper.UploadedFile{
			FieldName: "files[]",
			FileName:  "someJpgImg",
			File:      jpgImage,
		},
		helper.UploadedFile{
			FieldName: "files[]",
			FileName:  "somePngImg",
			File:      pngImage,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)
	assert.Len(t, filesResp.FilesToReturn, 2)

	assert.Regexp(t, `[\w]+/someJpgImg.jpg`, filesResp.FilesToReturn[0])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[0]))

	assert.Regexp(t, `[\w]+/somePngImg.png`, filesResp.FilesToReturn[1])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[1]))
}

func testTooBigDimension(t *testing.T) {
	img, err := helper.CreateImage(helper.ImageSpec{Format: "jpg", Width: 1000, Height: 600})
	assert.NoError(t, err)

	var filesResp filesResponse
	statusCode, err := makeTestingPost(
		&filesResp,
		helper.UploadedFile{
			FieldName: "files[]",
			FileName:  "someImage.jpg",
			File:      img,
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)
	assert.Len(t, filesResp.FilesToReturn, 1)

	if len(filesResp.FilesToReturn) == 0 {
		return
	}

	filePath := filepath.Join(helper.AssetsPath, filesResp.FilesToReturn[0])

	file, err := os.Open(filePath)
	assert.NoError(t, err)

	imgConfig, _, err := image.DecodeConfig(file)
	assert.NoError(t, err)
	assert.Equal(t, 833, imgConfig.Width)
	assert.Equal(t, 500, imgConfig.Height)
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
	err := saveImage(helper.AssetsPath, "lsls", "someImg.png", "png", 10, 10)
	assert.NoError(t, err)

	err = saveImage(
		helper.AssetsPath,
		filepath.Join("cache", "resized_image", "lsls", "someImg"),
		"5x5.png",
		"png",
		5,
		5,
	)
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

func testRead(t *testing.T) {
	err := saveImage(
		helper.AssetsPath,
		"imagesToRead",
		"someImg.png",
		"png",
		50,
		50,
	)
	assert.NoError(t, err)

	testClient := helper.NewTestClient()
	statusCode, body, err := testClient.MakeGet("http://localhost:9925/images/imagesToRead/someImg.png")
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)
	assertSameImage(
		t,
		filepath.Join(helper.AssetsPath, "imagesToRead", "someImg.png"),
		body,
	)
}

func assertSameImage(t *testing.T, sourceImgPath, body string) {
	sourceFile, err := os.Open(sourceImgPath)
	assert.NoError(t, err)
	defer sourceFile.Close()

	savedFileHash := md5.New()
	_, err = io.Copy(savedFileHash, sourceFile)
	assert.NoError(t, err)

	responseHash := md5.New()
	_, err = io.WriteString(responseHash, body)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%x", savedFileHash.Sum(nil)),
		fmt.Sprintf("%x", responseHash.Sum(nil)),
	)
}

func testGettingCachedResizedImage(t *testing.T) {
	err := saveImage(
		helper.AssetsPath,
		"imagesToResizeAndCache",
		"someImg.png",
		"png",
		500,
		250,
	)
	assert.NoError(t, err)

	filePath := filepath.Join(
		"cache",
		"resized_image",
		"imagesToResizeAndCache",
		"someImg",
	)

	err = saveImage(helper.AssetsPath, filePath, "100x100.png", "png", 500, 250)
	assert.NoError(t, err)

	testClient := helper.NewTestClient()
	statusCode, body, err := testClient.MakeGet(
		"http://localhost:9925/images/100x100/imagesToResizeAndCache/someImg.png",
	)

	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)

	bodyBuffer := bytes.NewBuffer([]byte(body))
	imgConfig, _, err := image.DecodeConfig(bodyBuffer)
	assert.NoError(t, err)

	// if we get an 100x100 image, it means the image was resized rather than using original cached 500x250 image
	assert.Equal(t, 500, imgConfig.Width)
	assert.Equal(t, 250, imgConfig.Height)
}

func testGettingResizedImage(t *testing.T) {
	err := saveImage(
		helper.AssetsPath,
		"imagesToResize",
		"someImg.png",
		"png",
		500,
		250,
	)
	assert.NoError(t, err)

	testCases := []struct {
		urlSize         string
		resizedFileName string
		expectedWidth   int
		expectedHeight  int
	}{
		{
			"100x",
			"100x.png",
			100,
			0,
		},
		{
			"x200",
			"x200.png",
			0,
			200,
		},
		{
			"300x150",
			"300x150.png",
			300,
			150,
		},
	}

	testClient := helper.NewTestClient()

	for _, testCase := range testCases {
		statusCode, body, err := testClient.MakeGet(
			fmt.Sprintf("http://localhost:9925/images/%s/imagesToResize/someImg.png", testCase.urlSize),
		)
		assert.NoError(t, err)
		assert.Equal(t, http2.StatusOK, statusCode)

		filePath := filepath.Join(
			helper.AssetsPath,
			"cache",
			"resized_image",
			"imagesToResize",
			"someImg",
			testCase.resizedFileName,
		)
		assert.FileExists(t, filePath)

		bodyBuffer := bytes.NewBuffer([]byte(body))
		imgConfig, _, err := image.DecodeConfig(bodyBuffer)
		assert.NoError(t, err)

		if testCase.expectedWidth > 0 {
			assert.Equal(t, testCase.expectedWidth, imgConfig.Width)
		}
		if testCase.expectedHeight > 0 {
			assert.Equal(t, testCase.expectedHeight, imgConfig.Height)
		}
	}
}

func testProxyMatch(t *testing.T) {
	err := saveImage(
		helper.ProxyAssetsPath,
		"imageToProxy",
		"someImg.jpg",
		"jpg",
		50,
		50,
	)
	assert.NoError(t, err)

	err = saveImage(
		helper.ProxyAssetsPath,
		filepath.Join("cache", "resized_image", "imageToProxy", "someImg"),
		"10x10.jpg",
		"jpg",
		10,
		10,
	)
	assert.NoError(t, err)

	testClient := helper.NewTestClient()
	statusCode, body, err := testClient.MakeGet("http://localhost:9925/images/imageToProxy/someImg.jpg")
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusOK, statusCode)

	assertSameImage(
		t,
		filepath.Join(helper.ProxyAssetsPath, "imageToProxy", "someImg.jpg"),
		body,
	)

	statusCodeResized, bodyResized, errResized := testClient.MakeGet(
		"http://localhost:9925/images/10x10/imageToProxy/someImg.jpg",
	)
	assert.NoError(t, errResized)
	assert.Equal(t, http2.StatusOK, statusCodeResized)

	assertSameImage(
		t,
		filepath.Join(helper.ProxyAssetsPath, "cache", "resized_image", "imageToProxy", "someImg", "10x10.jpg"),
		bodyResized,
	)
}

func testReadNonExistingImage(t *testing.T) {
	testClient := helper.NewTestClient()
	statusCode, body, err := testClient.MakeGet("http://localhost:9925/images/nonExistingImg/someImg.png")
	assert.NoError(t, err)
	assert.Equal(t, http2.StatusNotFound, statusCode)
	assert.Equal(t, "404 page not found\n", body)
}

func saveImage(rootPath, folderPath, imageName, format string, width, height int) error {
	img, err := helper.CreateImage(helper.ImageSpec{Format: format, Width: width, Height: height})
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(rootPath, folderPath), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(rootPath, folderPath, imageName))
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
	err := os.Setenv("PROXY_URL", "http://localhost:9926")
	errs.FailOnError(err)

	err = helper.PrepareFileServer(
		"media",
		helper.AssetsPath,
		map[string]string{
			"ASSETS_PATH":            helper.AssetsPath,
			"HOST":                   ":9925",
			"TOKEN_ISSUER":           "media-service-test",
			"TOKEN_SECRET":           "12345678",
			"URL_PREFIX":             "/images",
			"MAX_UPLOADED_FILE_MB":   "0.1",
			"HORIZ_MAX_IMAGE_HEIGHT": "500",
		},
	)
	errs.FailOnError(err)

	err = os.Unsetenv("PROXY_URL")
	errs.FailOnError(err)
	err = helper.PrepareFileServer(
		"proxy",
		helper.ProxyAssetsPath,
		map[string]string{
			"ASSETS_PATH": helper.ProxyAssetsPath,
			"HOST":        ":9926",
		},
	)
	errs.FailOnError(err)
}

func cleanup() {
	helper.ShutdownFileServers()
}
