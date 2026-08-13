package link

import (
	"context"
	"errors"
	"go-api/pkg/response"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidURL    = errors.New("invalid URL")
	ErrInvalidLinkID = errors.New("invalid link ID")
	ErrLinkNotFound  = errors.New("link not found")
	ErrHashCollision = errors.New("link hash collision")
)

func isHashCollision(err error) bool {
	var pgError *pgconn.PgError

	return errors.As(err, &pgError) &&
		pgError.Code == "23505" &&
		pgError.ConstraintName == "uq_links_hash"
}

func (handler *Handler) handleError(w http.ResponseWriter, req *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidURL):
		response.JsonError(
			w,
			http.StatusBadRequest,
			"INVALID_URL",
			"URL is invalid",
		)

	case errors.Is(err, ErrInvalidLinkID):
		response.JsonError(
			w,
			http.StatusBadRequest,
			"INVALID_LINK_ID",
			"Link ID is invalid",
		)

	case errors.Is(err, ErrLinkNotFound):
		response.JsonError(
			w,
			http.StatusNotFound,
			"LINK_NOT_FOUND",
			"Link not found",
		)

	case errors.Is(err, context.DeadlineExceeded):
		handler.logger.WarnContext(
			req.Context(),
			"request deadline exceeded",
			"method", req.Method,
			"path", req.URL.Path,
		)
		response.JsonError(
			w,
			http.StatusGatewayTimeout,
			"REQUEST_TIMEOUT",
			"Request took too long",
		)

	case errors.Is(err, ErrHashCollision):
		handler.logger.WarnContext(
			req.Context(),
			"link hash collision",
			"method", req.Method,
			"path", req.URL.Path,
		)
		response.JsonError(
			w,
			http.StatusConflict,
			"HASH_COLLISION",
			"Link hash already exists",
		)

	case errors.Is(err, context.Canceled):
		return
	default:
		handler.logger.ErrorContext(
			req.Context(),
			err.Error(),
			"error", err,
			"method", req.Method,
			"path", req.URL.Path,
		)
		response.JsonError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal server error",
		)
	}
}
