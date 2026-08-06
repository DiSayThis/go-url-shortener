package link

import (
	"context"
	"fmt"
	"go-api/internal/database"
	"go-api/pkg/request"
	"go-api/pkg/response"
	"net/http"
)

type LinkHandlerDeps struct {
	Service LinkService
}
type LinkService interface {
	Create(
		ctx context.Context,
		rawURL string,
	) (*database.Link, error)
}

type Handler struct {
	Service LinkService
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {
	handler := &Handler{Service: deps.Service}

	router.HandleFunc("POST /link", handler.create)
	router.HandleFunc("PATCH /link/{id}", handler.update)
	router.HandleFunc("DELETE /link/{id}", handler.delete)
	router.HandleFunc("GET /{hash}", handler.get)
}

type CreateLinkRequest struct {
	Url string `json:"url" validate:"required,url"`
}

func (handler *Handler) create(w http.ResponseWriter, req *http.Request) {
	payload, err := request.HandleBody[CreateLinkRequest](w, req)
	if err != nil {
		return
	}
	result, err := handler.Service.Create(
		req.Context(),
		payload.Url,
	)
	if err != nil {
		response.JsonResponse(w, map[string]string{"error": "Failed to create link: " + err.Error()}, http.StatusInternalServerError)
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

func (handler *Handler) get(w http.ResponseWriter, req *http.Request) {
}
