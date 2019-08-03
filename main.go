package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-service/http"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	action := *flag.String("Action", "server", "Use server or cli")

	switch action {
	case "server":
		serverRunner := http.NewServerRunner()
		srv, err := serverRunner.Run()
		errs.FailOnError(err)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, os.Kill)
		<-stop

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		err = srv.Shutdown(ctx)
		errs.FailOnError(err)

		break
	case "cli":
		fmt.Println("Todo cli")
		break
	default:
		log.Panicf("Invalid action %s\n", action)

	}
}
