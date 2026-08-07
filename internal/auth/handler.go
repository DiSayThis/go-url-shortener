package auth

import (
	"fmt"
	"go-api/configs"
	"go-api/pkg/request"
	"go-api/pkg/response"
	"net/http"
)

type AuthHandlerDeps struct {
	*configs.Config
}
type AuthHandler struct {
	*configs.Config
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{Config: deps.Config}
	router.HandleFunc("POST /auth/login", handler.login)
	router.HandleFunc("POST /auth/register", handler.register)
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	payload, err := request.HandleBody[LoginRequest](w, req)
	if err != nil {
		return
	}
	fmt.Println("login", payload)
	data := LoginResponse{
		Token: handler.Config.Auth.Secret,
	}
	response.JsonResponse(w, data, http.StatusOK)
}

func (handler *AuthHandler) register(w http.ResponseWriter, req *http.Request) {
	payload, err := request.HandleBody[RegisterRequest](w, req)
	if err != nil {
		return
	}
	fmt.Println("register", payload)
	data := RegisterResponse{
		Token: handler.Config.Auth.Secret,
	}
	response.JsonResponse(w, data, http.StatusCreated)
}
