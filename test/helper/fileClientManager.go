package helper

import (
	"bytes"
	"fmt"
	"github.com/breathbath/media-service/authentication"
	"io"
	"io/ioutil"
	"mime/multipart"
	http2 "net/http"
)

type TestClient struct {
	body        *bytes.Buffer
	contentType string
}

type UploadedFile struct {
	FieldName string
	FileName  string
	File      io.Reader
}

func NewTestClient() *TestClient {
	return &TestClient{body: &bytes.Buffer{}}
}

func (tc *TestClient) AddFiles(files ...UploadedFile) error {
	writer := multipart.NewWriter(tc.body)
	for _, uploadedFile := range files {
		part, err := writer.CreateFormFile(uploadedFile.FieldName, uploadedFile.FileName)
		if err != nil {
			return err
		}

		_, err = io.Copy(part, uploadedFile.File)
		if err != nil {
			return err
		}
	}

	tc.contentType = writer.FormDataContentType()

	err := writer.Close()
	if err != nil {
		return err
	}

	return nil
}

func (tc *TestClient) GenerateValidToken() (string, error) {
	jwtManager, err := authentication.NewJwtManager()
	if err != nil {
		return "", err
	}

	return jwtManager.GenerateToken("test")
}

func (tc *TestClient) MakePost(token, url string) (statusCode int, body string, err error) {
	r, _ := http2.NewRequest("POST", url, tc.body)
	if tc.contentType != "" {
		r.Header.Add("Content-Type", tc.contentType)
	}

	if token != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http2.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return 0, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, string(respBody), nil
}

func (tc *TestClient) MakeDelete(token, url string) (statusCode int, err error) {
	r, _ := http2.NewRequest("DELETE", url, &bytes.Buffer{})

	if token != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http2.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
