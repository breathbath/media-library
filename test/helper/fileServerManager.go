package helper

import (
	"context"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-service/http"
	http2 "net/http"
	"os"
)

var srv *http2.Server
const AssetsPath = "/tmp/assets"

func PrepareFileServer() error {
	var err error

	err = os.MkdirAll(AssetsPath, os.ModePerm)
	if err != nil {
		return err
	}

	err = SetEnvs(map[string]string{
		"ASSETS_PATH": AssetsPath,
		"HOST": ":9925",
		"TOKEN_ISSUER": "media-service-test",
		"TOKEN_SECRET": "12345678",
		"URL_PREFIX": "/images",
		"MAX_UPLOADED_FILE_MB": "0.1",
		"HORIZ_MAX_IMAGE_HEIGHT": "500",
	})
	if err != nil {
		return err
	}

	serverRunner := http.NewServerRunner()
	srv, err = serverRunner.Run()
	if err != nil {
		return err
	}
	return nil
}

func ShutdownFileServer() {
	err := srv.Shutdown(context.Background())
	errs.FailOnError(err)

	err = os.RemoveAll(AssetsPath)
	errs.FailOnError(err)
}