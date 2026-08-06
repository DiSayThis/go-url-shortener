package link

import (
	"go-api/configs"
	"net/http"
)

type LinkHandlerDeps struct {
	*configs.Config
}
type LinkHandler struct {
	*configs.Config
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {
	handler := &LinkHandler{Config: deps.Config}

	router.HandleFunc("POST /link/create", handler.create)
	router.HandleFunc("PATCH /link/{id}", handler.update)
	router.HandleFunc("DELETE /link/{id}", handler.delete)
	router.HandleFunc("GET /{alias}", handler.get)
}

func (handler *LinkHandler) create(w http.ResponseWriter, req *http.Request) {
}

func (handler *LinkHandler) update(w http.ResponseWriter, req *http.Request) {
}

func (handler *LinkHandler) delete(w http.ResponseWriter, req *http.Request) {
}

func (handler *LinkHandler) get(w http.ResponseWriter, req *http.Request) {
}
