package http

import (
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-service/authentication"
	"github.com/breathbath/media-service/files"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
)

type ServerRunner struct {
}

func NewServerRunner() ServerRunner {
	return ServerRunner{}
}

func (sr ServerRunner) Run() error {
	host, err := env.ReadEnvOrError("HOST")
	if err != nil {
		return err
	}

	assetsPath, err := env.ReadEnvOrError("ASSETS_PATH")
	if err != nil {
		return err
	}

	urlPrefix := env.ReadEnv("URL_PREFIX", "/media/images/")

	recoveryHandler := negroni.NewRecovery()
	recoveryHandler.Formatter = &PanicFormatter{}

	jwtManager, err := authentication.NewJwtManager()
	if err != nil {
		return err
	}
	authHandler := authentication.NewAuthHandlerProvider(jwtManager)

	serverHandler := negroni.New(
		negroni.NewLogger(),
		recoveryHandler,
		negroni.HandlerFunc(authHandler.GetHandlerFunc()),
	)

	router := mux.NewRouter()

	fileServerHandler := files.NewFileServer(assetsPath)
	router.PathPrefix(urlPrefix).Handler(http.StripPrefix(urlPrefix, fileServerHandler))

	serverHandler.UseHandler(router)

	io.OutputInfo("", "Starting server at %s behind official host %s", host, host)

	return http.ListenAndServe(host, serverHandler)
}

