package test

import (
	"encoding/json"
	"fmt"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-service/test/helper"
	"github.com/stretchr/testify/assert"
	http2 "net/http"
	"path/filepath"
	"testing"
)

func TestPostHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	setup()
	defer cleanup()
	t.Run("testImageSaved", testImageSaved)
}

func testImageSaved(t *testing.T) {
	jpgImage, err := helper.CreateImage("jpg")
	assert.NoError(t, err)

	pngImage, err := helper.CreateImage("png")
	assert.NoError(t, err)

	testClient := helper.NewTestClient()
	err = testClient.AddFiles(
		helper.UploadedFile{
			FieldName:"files",
			FileName: "someImage.jpg",
			File: jpgImage,
		},
		helper.UploadedFile{
			FieldName:"files",
			FileName: "someImage.png",
			File: pngImage,
		},
	)
	assert.NoError(t, err)

	validToken, err := testClient.GenerateValidToken()
	assert.NoError(t, err)

	statusCode, body, err := testClient.MakePost(validToken, "http://localhost:9925/images")
	assert.NoError(t, err)
	fmt.Printf("Received response '%s'\n", body)

	assert.Equal(t, http2.StatusOK, statusCode)

	var fileToReturn struct {FilesToReturn []string `json:"filepathes"`}
	err = json.Unmarshal([]byte(body), &fileToReturn)
	assert.NoError(t, err)

	assert.Len(t, fileToReturn.FilesToReturn, 2)

	assert.Regexp(t, `[\w]+/someImage.jpg`, fileToReturn.FilesToReturn[0])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, fileToReturn.FilesToReturn[0]))

	assert.Regexp(t, `[\w]+/someImage.png`, fileToReturn.FilesToReturn[1])
	assert.FileExists(t, filepath.Join(helper.AssetsPath, fileToReturn.FilesToReturn[1]))
}

func setup() {
	err := helper.PrepareFileServer()
	errs.FailOnError(err)
}

func cleanup() {
	helper.ShutdownFileServer()
}
