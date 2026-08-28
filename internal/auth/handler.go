package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

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
	router.HandleFunc("POST /auth/logout", handler.logout)
	router.HandleFunc("POST /auth/sessions", handler.sessions)
	router.HandleFunc("POST /auth/sessions/{familyID}", handler.sessionsByFamilyId)
	router.HandleFunc("POST /auth/logout-all", handler.logoutAll)
}

func preventAuthResponseCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

func (handler *AuthHandler) login(w http.ResponseWriter, req *http.Request) {
	preventAuthResponseCaching(w)
	req.Body = http.MaxBytesReader(w, req.Body, 64<<10)
	body, err := request.HandleBody[LoginRequest](w, req)
	if err != nil {
		return
	}
	result, err := handler.Service.Login(req.Context(), LoginInput{
		Email:     body.Email,
		Password:  body.Password,
		UserAgent: req.UserAgent(),
		CreatedIP: requestRemoteIP(req),
	})
	if err != nil {
		handler.handleError(w, req, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    result.RefreshToken,
		Path:     refreshCookiePath,
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
	response.JsonResponse(w, RegisterResponse{
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}, http.StatusCreated)
}

func (handler *AuthHandler) refresh(w http.ResponseWriter, req *http.Request) {
	preventAuthResponseCaching(w)
	cookie, err := req.Cookie("refresh_token")
	if errors.Is(err, http.ErrNoCookie) {
		handler.handleError(w, req, ErrInvalidRefreshSession)
		return
	}
	if err != nil {
		handler.handleError(w, req, ErrInvalidRefreshSession)
		return
	}
	rawToken := cookie.Value
	if rawToken == "" {
		handler.handleError(w, req, ErrInvalidRefreshSession)
		return
	}
	result, err := handler.Service.Refresh(
		req.Context(),
		RefreshInput{
			RefreshToken: rawToken,
			UserAgent:    req.UserAgent(),
			ClientIP:     requestRemoteIP(req),
		},
	)
	if err != nil {
		handler.handleError(w, req, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    result.RefreshToken,
		Path:     refreshCookiePath,
		HttpOnly: true,
		Secure:   handler.RefreshCookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.RefreshTokenExpiresAt,
		MaxAge:   int(result.RefreshTokenTTL.Seconds()),
	})
	response.JsonResponse(
		w,
		RefreshResponse{
			AccessToken: result.AccessToken,
			TokenType:   "Bearer",
			ExpiresAt:   result.AccessTokenExpiresAt,
		},
		http.StatusOK,
	)
}

func (handler *AuthHandler) logout(w http.ResponseWriter, req *http.Request) {
}
func (handler *AuthHandler) sessions(w http.ResponseWriter, req *http.Request) {
}
func (handler *AuthHandler) sessionsByFamilyId(w http.ResponseWriter, req *http.Request) {
}
func (handler *AuthHandler) logoutAll(w http.ResponseWriter, req *http.Request) {
}

const (
	refreshCookieName = "refresh_token"
	refreshCookiePath = "/auth"
)

func (handler *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		HttpOnly: true,
		Secure:   handler.RefreshCookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}
