package link

import (
	"fmt"
	"time"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateLinkRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type UpdateLinkRequest struct {
	URL  string `json:"url" validate:"required,url"`
	Hash string `json:"hash"`
}

type LinkResponse struct {
	PublicID  string     `json:"id"`
	URL       string     `json:"url"`
	Hash      string     `json:"hash"`
	Title     *string    `json:"title"`
	IsActive  bool       `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func newLinkResponse(link *database.Link) LinkResponse {
	return LinkResponse{
		PublicID:  formatUUID(link.PublicID),
		URL:       link.Url,
		Hash:      link.Hash,
		Title:     textPointer(link.Title),
		IsActive:  link.IsActive,
		ExpiresAt: timePointer(link.ExpiresAt),
		CreatedAt: link.CreatedAt.Time,
		UpdatedAt: link.UpdatedAt.Time,
	}
}

func formatUUID(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}

	bytes := value.Bytes
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	)
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
