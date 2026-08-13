package auth

import (
	"go-api/pkg/request"
	"go-api/pkg/response"
	"log/slog"
	"net/http"
)

type AuthHandlerDeps struct {
	Service AuthService
	Logger  *slog.Logger
}
type AuthHandler struct {
	Service AuthService
	Logger  *slog.Logger
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{}
	router.HandleFunc("POST /auth/login", handler.login)
	router.HandleFunc("POST /auth/register", handler.register)
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	_, err := request.HandleBody[LoginRequest](w, req)
	if err != nil {
		return
	}
	response.JsonResponse(w, nil, http.StatusOK)
}

func (handler *AuthHandler) register(w http.ResponseWriter, req *http.Request) {
	_, err := request.HandleBody[RegisterRequest](w, req)
	if err != nil {
		return
	}
	response.JsonResponse(w, nil, http.StatusCreated)
}
