package link

import (
	"context"
	"errors"
	"fmt"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type LinkRepository interface {
	Create(ctx context.Context, userID int64, link *database.Link) (*database.Link, error)
	GetByHash(ctx context.Context, hash string) (*database.Link, error)
	Update(ctx context.Context, userID int64, publicID pgtype.UUID, url string) (*database.Link, error)
	UpdateWithHash(ctx context.Context, userID int64, publicID pgtype.UUID, url, hash string) (*database.Link, error)
	Delete(ctx context.Context, userID int64, publicID pgtype.UUID) error
}

type Repository struct {
	queries database.Querier
}

func NewLinkRepository(queries database.Querier) *Repository {
	return &Repository{queries: queries}
}

func (repo *Repository) Create(ctx context.Context, userID int64, link *database.Link) (*database.Link, error) {
	createdLink, err := repo.queries.CreateLink(ctx, database.CreateLinkParams{
		UserID: userID,
		Url:    link.Url,
		Hash:   link.Hash,
	})
	if err != nil {
		if isHashCollision(err) {
			return nil, ErrHashCollision
		}
		return nil, fmt.Errorf("create link: %w", err)
	}

	return &createdLink, nil
}

func (repo *Repository) GetByHash(ctx context.Context, hash string) (*database.Link, error) {
	foundLink, err := repo.queries.GetLinkByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("select link by hash %q: %w", hash, err)
	}

	return &foundLink, nil
}

func (repo *Repository) Update(
	ctx context.Context,
	userID int64,
	publicID pgtype.UUID,
	url string,
) (*database.Link, error) {
	updatedLink, err := repo.queries.UpdateLinkURL(ctx, database.UpdateLinkURLParams{
		Url:      url,
		PublicID: publicID,
		UserID:   userID,
	})
	if err != nil {
		return nil, handleLinkMutationError("update link", err)
	}

	return &updatedLink, nil
}

func (repo *Repository) UpdateWithHash(
	ctx context.Context,
	userID int64,
	publicID pgtype.UUID,
	url string,
	hash string,
) (*database.Link, error) {
	updatedLink, err := repo.queries.UpdateLinkUrlAndHash(ctx, database.UpdateLinkUrlAndHashParams{
		Url:      url,
		Hash:     hash,
		PublicID: publicID,
		UserID:   userID,
	})
	if err != nil {
		return nil, handleLinkMutationError("update link and hash", err)
	}

	return &updatedLink, nil
}

func (repo *Repository) Delete(ctx context.Context, userID int64, publicID pgtype.UUID) error {
	_, err := repo.queries.SoftDeleteLink(ctx, database.SoftDeleteLinkParams{
		PublicID: publicID,
		UserID:   userID,
	})
	if err != nil {
		return handleLinkMutationError("delete link", err)
	}

	return nil
}

func handleLinkMutationError(operation string, err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrLinkNotFound
	case isHashCollision(err):
		return ErrHashCollision
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
