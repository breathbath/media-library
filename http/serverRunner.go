package http

import (
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-library/assets"
	"github.com/breathbath/media-library/authentication"
	"github.com/breathbath/media-library/fileSystem"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type ServerRunner struct {
}

func NewServerRunner() ServerRunner {
	return ServerRunner{}
}

func (sr ServerRunner) Run() (*http.Server, error) {
	host, err := env.ReadEnvOrError("HOST")
	if err != nil {
		return nil, err
	}

	assetsPath, err := env.ReadEnvOrError("ASSETS_PATH")
	if err != nil {
		return nil, err
	}

	urlPrefix := env.ReadEnv("URL_PREFIX", "/media/images/")

	recoveryHandler := negroni.NewRecovery()
	recoveryHandler.Formatter = &PanicFormatter{}

	jwtManager, err := authentication.NewJwtManager()
	if err != nil {
		return nil, err
	}
	authHandler := authentication.NewAuthHandlerProvider(jwtManager)

	serverHandler := negroni.New(
		negroni.NewLogger(),
		recoveryHandler,
		negroni.HandlerFunc(authHandler.GetHandlerFunc()),
	)

	router := mux.NewRouter()

	fileSystemHandler := fileSystem.LocalFileSystemManager{AssetsPath: assetsPath}

	fileSystemManager := assets.NewImageReadHandler(fileSystemHandler)
	fileServerHandler := http.FileServer(fileSystemManager)
	router.PathPrefix(urlPrefix).Handler(http.StripPrefix(urlPrefix, fileServerHandler)).Methods(http.MethodGet)

	imageDeleteHandler := assets.ImageDeleteHandler{
		FileSystemManager: fileSystemHandler,
	}
	router.HandleFunc(strings.TrimRight(urlPrefix, "/") + "/{folder}/{image}", imageDeleteHandler.HandleDelete).Methods(http.MethodDelete)

	imageSaver := assets.ImageSaver{FileSystemHandler: fileSystemHandler}
	postHandler := assets.ImagePostHandler{ImageSaver: imageSaver}
	router.HandleFunc(urlPrefix, postHandler.HandlePost).Methods(http.MethodPost)

	serverHandler.UseHandler(router)

	io.OutputInfo("", "Starting server at %s behind official host %s", host, host)

	srv := &http.Server{Addr: host, Handler: serverHandler}

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			io.OutputError(err, "", "")
		}
	}()

	return srv, nil
}

