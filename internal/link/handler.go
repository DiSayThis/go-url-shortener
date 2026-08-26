package link

import (
	"log/slog"
	"net/http"

	"go-api/internal/auth"
	"go-api/pkg/middleware"
	"go-api/pkg/request"
	"go-api/pkg/response"
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

	// RequireAuth гарантирует, что защищённые handlers получат Principal.
	router.Handle("POST /links", middleware.RequireAuth(handler.create))
	router.Handle("PATCH /links/{publicID}", middleware.RequireAuth(handler.update))
	router.Handle("DELETE /links/{publicID}", middleware.RequireAuth(handler.delete))

	// Redirect публичный, поэтому middleware здесь не нужен.
	router.HandleFunc("GET /{hash}", handler.goTo)
}

func (handler *Handler) create(w http.ResponseWriter, req *http.Request) {
	principal, ok := auth.RequirePrincipal(w, req)
	if !ok {
		return
	}
	payload, err := request.HandleBody[CreateLinkRequest](w, req)
	if err != nil {
		return
	}

	// UserID берётся из проверенного Principal, а не из JSON клиента.
	result, err := handler.service.Create(req.Context(), principal.UserID, payload.URL)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}

	response.JsonResponse(w, newLinkResponse(result), http.StatusCreated)
}

func (handler *Handler) update(w http.ResponseWriter, req *http.Request) {
	principal, ok := auth.RequirePrincipal(w, req)
	if !ok {
		return
	}
	payload, err := request.HandleBody[UpdateLinkRequest](w, req)
	if err != nil {
		return
	}

	result, err := handler.service.Update(
		req.Context(),
		principal.UserID,
		req.PathValue("publicID"),
		payload.URL,
		payload.Hash,
	)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}

	response.JsonResponse(w, newLinkResponse(result), http.StatusOK)
}

func (handler *Handler) delete(w http.ResponseWriter, req *http.Request) {
	principal, ok := auth.RequirePrincipal(w, req)
	if !ok {
		return
	}
	if err := handler.service.Delete(
		req.Context(),
		principal.UserID,
		req.PathValue("publicID"),
	); err != nil {
		handler.handleError(w, req, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) goTo(w http.ResponseWriter, req *http.Request) {
	foundLink, err := handler.service.GetByHash(req.Context(), req.PathValue("hash"))
	if err != nil {
		handler.handleError(w, req, err)
		return
	}

	http.Redirect(w, req, foundLink.Url, http.StatusTemporaryRedirect)
}
