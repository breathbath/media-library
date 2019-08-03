package http

import (
	"net/http"
)

type NotFoundHandler struct {
}

func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

func (nfh *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
