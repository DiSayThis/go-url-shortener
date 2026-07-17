package auth

import (
	"encoding/json"
	"fmt"
	"go-api/configs"
	"go-api/pkg/res"
	"net/http"
	"regexp"
)

type AuthHandlerDeps struct {
	*configs.Config
}
type AuthHandler struct {
	*configs.Config
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{Config: deps.Config}
	prefix := "/auth"
	router.HandleFunc("POST "+prefix+"/login", handler.login)
	router.HandleFunc("POST "+prefix+"/register", handler.register)
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	var payload LoginRequest
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		res.JsonResponse(w, map[string]string{"error": "Invalid request body: " + err.Error()}, http.StatusBadRequest)
		return
	}
	emailRegex, _ := regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if payload.Email == "" {
		res.JsonResponse(w, map[string]string{"error": "Email is required"}, http.StatusBadRequest)
		return
	}
	if !emailRegex.MatchString(payload.Email) {
		res.JsonResponse(w, map[string]string{"error": "Invalid email format"}, http.StatusBadRequest)
		return
	}
	if payload.Password == "" {
		res.JsonResponse(w, map[string]string{"error": "Password is required"}, http.StatusBadRequest)
		return
	}
	fmt.Println("login", payload.Email, payload.Password)
	data := LoginResponse{
		Token: handler.Config.Auth.Secret,
	}
	res.JsonResponse(w, data, http.StatusOK)
}

func (handler *AuthHandler) register(w http.ResponseWriter, req *http.Request) {
	fmt.Println("register")
}
