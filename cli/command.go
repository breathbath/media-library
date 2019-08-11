package cli

import (
	"context"
	"fmt"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-service/authentication"
	"github.com/breathbath/media-service/http"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"time"
)

func Start() {
	var (
		app = kingpin.New("media", "Media service")

		tokenGenerator   = app.Command("token", "Generates new token")
		appName = tokenGenerator.Arg("app", "Application name").Required().String()

		mediaServer   = app.Command("server", "Starts media server")
	)

	kingpin.Version("1.0.0")
	parsedCliInput := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch parsedCliInput {
	case tokenGenerator.FullCommand():
		jwtManager, err := authentication.NewJwtManager()
		errs.FailOnError(err)

		token, err := jwtManager.GenerateToken(*appName)
		errs.FailOnError(err)

		fmt.Println(token)
		break
	case mediaServer.FullCommand():
		serverRunner := http.NewServerRunner()
		srv, err := serverRunner.Run()
		errs.FailOnError(err)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, os.Kill)
		<-stop

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		err = srv.Shutdown(ctx)
		errs.FailOnError(err)
		cancelFunc()

		break
	}
}
