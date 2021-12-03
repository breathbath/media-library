package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-library/authentication"
	"github.com/breathbath/media-library/http"
	"gopkg.in/alecthomas/kingpin.v2"
)

func Start() {
	var (
		app = kingpin.New("media", "Media service")

		tokenGenerator = app.Command("token", "Generates new token")
		appName        = tokenGenerator.Arg("app", "Application name").Required().String()

		mediaServer = app.Command("server", "Starts media server")
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
	case mediaServer.FullCommand():
		serverRunner := http.NewServerRunner()
		srv, err := serverRunner.Run()
		errs.FailOnError(err)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop

		const startingTimeout = 5 * time.Second
		ctx, cancelFunc := context.WithTimeout(context.Background(), startingTimeout)
		err = srv.Shutdown(ctx)
		errs.FailOnError(err)
		cancelFunc()
	}
}
