package helper

import (
	"context"
	http2 "net/http"
	"os"

	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-library/http"
)

var srvs map[string]*http2.Server
var assetsPaths []string

func init() {
	srvs = make(map[string]*http2.Server)
	assetsPaths = []string{}
}

const AssetsPath = "/tmp/assets"
const ProxyAssetsPath = "/tmp/proxy"

func PrepareFileServer(name, assetsPath string, envs map[string]string) error {
	var err error

	err = os.MkdirAll(assetsPath, os.ModePerm)
	if err != nil {
		return err
	}

	err = SetEnvs(envs)
	if err != nil {
		return err
	}

	serverRunner := http.NewServerRunner()
	srvs[name], err = serverRunner.Run()
	if err != nil {
		return err
	}

	assetsPaths = append(assetsPaths, assetsPath)

	return nil
}

func ShutdownFileServers() {
	errc := errs.NewErrorContainer()
	for _, srv := range srvs {
		err := srv.Shutdown(context.Background())
		if err != nil {
			errc.AddError(err)
		}
	}
	errs.FailOnError(errc.Result(" "))

	for _, assetPath := range assetsPaths {
		err := os.RemoveAll(assetPath)
		if err != nil {
			errc.AddError(err)
		}
	}
	errs.FailOnError(errc.Result(" "))
}
