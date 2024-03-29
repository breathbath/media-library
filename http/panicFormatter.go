package http

import (
	"fmt"
	"net/http"

	"github.com/breathbath/go_utils/utils/io"
	"github.com/codegangsta/negroni"
)

type PanicFormatter struct{}

func (pf *PanicFormatter) FormatPanicError(rw http.ResponseWriter, r *http.Request, infos *negroni.PanicInformation) {
	panicContext := infos.RecoveredPanic
	recoveredError, isError := panicContext.(error)

	if !isError {
		recoveredError = fmt.Errorf("%v", panicContext)
	}

	io.OutputError(recoveredError, "Http Error", "Unexpected panic error")

	rw.WriteHeader(http.StatusInternalServerError)
}
