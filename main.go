package main

import (
	"flag"
	"fmt"
	"github.com/breathbath/go_utils/utils/errs"
	"github.com/breathbath/media-service/http"
	"log"
)

func main() {
	action := *flag.String("Action", "server", "Use server or cli")

	switch action {
	case "server":
		server := http.NewServerRunner()
		err := server.Run()
		errs.FailOnError(err)
		break
	case "cli":
		fmt.Println("Todo cli")
		break
	default:
		log.Panicf("Invalid action %s\n", action)

	}
}
