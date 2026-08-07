package link

import (
	"fmt"
	"go-api/pkg/request"
	"go-api/pkg/response"
	"log/slog"
	"net/http"
)

type LinkHandlerDeps struct {
	Service LinkService
	Logger  *slog.Logger
}

type Handler struct {
	service LinkService
	logger  *slog.Logger
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	handler := &Handler{
		service: deps.Service,
		logger:  logger.With("component", "link_handler"),
	}

	router.HandleFunc("POST /link", handler.create)
	router.HandleFunc("PATCH /link/{id}", handler.update)
	router.HandleFunc("DELETE /link/{id}", handler.delete)
	router.HandleFunc("GET /{hash}", handler.goTo)
}

type CreateLinkRequest struct {
	URL string `json:"url" validate:"required,url"`
}

func (handler *Handler) create(w http.ResponseWriter, req *http.Request) {
	payload, err := request.HandleBody[CreateLinkRequest](w, req)
	if err != nil {
		return
	}
	result, err := handler.service.Create(
		req.Context(),
		payload.URL,
	)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}
	response.JsonResponse(w, result, http.StatusCreated)
}

func (handler *Handler) update(w http.ResponseWriter, req *http.Request) {
}

func (handler *Handler) delete(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	fmt.Println(id)
}

func (handler *Handler) goTo(w http.ResponseWriter, req *http.Request) {
	hash := req.PathValue("hash")
	link, err := handler.service.GetByHash(req.Context(), hash)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}
	http.Redirect(w, req, link.Url, http.StatusTemporaryRedirect)
}
