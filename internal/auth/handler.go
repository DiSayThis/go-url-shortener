package auth

import (
	"log/slog"
	"net/http"

	"go-api/pkg/request"
	"go-api/pkg/response"
)

type AuthHandlerDeps struct {
	Service             AuthService
	Logger              *slog.Logger
	RefreshCookieSecure bool
}

type AuthHandler struct {
	Service             AuthService
	Logger              *slog.Logger
	RefreshCookieSecure bool
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	handler := &AuthHandler{
		Service:             deps.Service,
		Logger:              logger.With("component", "auth_handler"),
		RefreshCookieSecure: deps.RefreshCookieSecure,
	}

	router.HandleFunc("POST /auth/login", handler.login)
	router.HandleFunc("POST /auth/register", handler.register)
	router.HandleFunc("POST /auth/refresh", handler.refresh)
	router.HandleFunc("POST /auth/logout", handler.register)
	router.HandleFunc("POST /auth/sessions", handler.register)
	router.HandleFunc("POST /auth/sessions/{familyID}", handler.register)
	router.HandleFunc("POST /auth/logout-all", handler.register)
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	body, err := request.HandleBody[LoginRequest](w, req)
	if err != nil {
		return
	}
	result, err := handler.Service.Login(req.Context(), LoginInput{
		Email:     body.Email,
		Password:  body.Password,
		UserAgent: req.UserAgent(),
	})
	if err != nil {
		handler.handleError(w, req, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   handler.RefreshCookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.RefreshTokenExpiresAt,
		MaxAge:   int(result.RefreshTokenTTL.Seconds()),
	})

	responseBody := LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   result.AccessTokenExpiresAt,
		User: UserResponse{
			PublicID:    result.User.PublicID.String(),
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
			Role:        result.User.Role,
		},
	}

	response.JsonResponse(w, responseBody, http.StatusOK)
}

func (handler *AuthHandler) register(w http.ResponseWriter, req *http.Request) {
	body, err := request.HandleBody[RegisterRequest](w, req)
	if err != nil {
		return
	}
	input := RegisterInput{
		Email:       body.Email,
		DisplayName: body.Name,
		Password:    body.Password}
	user, err := handler.Service.Register(req.Context(), input)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}
	response.JsonResponse(w, user, http.StatusCreated)
}

func (handler *AuthHandler) refresh(w http.ResponseWriter, req *http.Request) {

}
