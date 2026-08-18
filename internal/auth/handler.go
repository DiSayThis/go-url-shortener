package auth

import (
	"fmt"
	"go-api/pkg/request"
	"go-api/pkg/response"
	"log/slog"
	"net/http"
)

type AuthHandlerDeps struct {
	Service      AuthService
	AccessTokens AccessTokenService
	Logger       *slog.Logger
}

type AuthHandler struct {
	Service      AuthService
	AccessTokens AccessTokenService
	Logger       *slog.Logger
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	handler := &AuthHandler{
		Service:      deps.Service,
		AccessTokens: deps.AccessTokens,
		Logger:       logger.With("component", "auth_handler"),
	}

	router.HandleFunc("POST /auth/login", handler.login)
	router.HandleFunc("POST /auth/register", handler.register)
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	body, err := request.HandleBody[LoginRequest](w, req)
	if err != nil {
		return
	}
	user, err := handler.Service.Authenticate(req.Context(), body.Email, body.Password)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}
	if !user.PublicID.Valid {
		handler.handleError(w, req,
			fmt.Errorf("authenticated user has invalid public ID"),
		)
		return
	}
	sessionID, err := generateTokenID()
	if err != nil {
		handler.handleError(
			w,
			req,
			fmt.Errorf("generate session ID: %w", err),
		)
		return
	}

	issuedToken, err := handler.AccessTokens.Issue(AccessTokenInput{
		UserID:    user.ID,
		PublicID:  user.PublicID.String(),
		Role:      user.Role,
		SessionID: sessionID,
		Scopes:    nil,
	})
	if err != nil {
		handler.handleError(
			w,
			req,
			fmt.Errorf("issue access token: %w", err),
		)
		return
	}

	result := LoginResponse{
		AccessToken: issuedToken.Token,
		TokenType:   "Bearer",
		ExpiresAt:   issuedToken.ExpiresAt,
		User: UserResponse{
			PublicID:    user.PublicID.String(),
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        user.Role,
		},
	}

	response.JsonResponse(w, result, http.StatusOK)
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
	fmt.Println(user)
	response.JsonResponse(w, user, http.StatusCreated)
}
