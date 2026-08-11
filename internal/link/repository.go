package link

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"

	"github.com/jackc/pgx/v5"
)

type LinkRepository interface {
	Create(ctx context.Context, link *database.Link) (*database.Link, error)
	GetByHash(ctx context.Context, hash string) (*database.Link, error)
	Update(ctx context.Context, link *database.Link) (*database.Link, error)
	UpdateWithHash(ctx context.Context, link *database.Link) (*database.Link, error)
	DeleteById(ctx context.Context, id int64) (int64, error)
}

type Repository struct {
	queries database.Querier
}

func NewLinkRepository(queries database.Querier) *Repository {
	return &Repository{
		queries: queries,
	}
}

func (repo *Repository) Create(ctx context.Context, link *database.Link) (*database.Link, error) {
	createdLink, err := repo.queries.CreateLink(
		ctx,
		database.CreateLinkParams{
			Url:  link.Url,
			Hash: link.Hash,
		},
	)
	if err != nil {
		if isHashCollision(err) {
			return nil, ErrHashCollision
		}
		return nil, fmt.Errorf("create link: %w", err)
	}

	return &createdLink, nil
}

func (repo *Repository) GetByHash(ctx context.Context, hash string) (*database.Link, error) {
	link, err := repo.queries.GetLinkByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("select link by hash %q: %w", hash, err)
	}
	return &link, nil
}

func (repo *Repository) Update(ctx context.Context, link *database.Link) (*database.Link, error) {
	updatedLink, err := repo.queries.UpdateLinkURL(ctx, database.UpdateLinkURLParams{ID: link.ID, Url: link.Url})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("update link: %w", err)
	}
	return &updatedLink, nil
}

func (repo *Repository) UpdateWithHash(ctx context.Context, link *database.Link) (*database.Link, error) {
	updatedLink, err := repo.queries.UpdateLinkUrlAndHash(ctx, database.UpdateLinkUrlAndHashParams{ID: link.ID, Url: link.Url, Hash: link.Hash})
	if err != nil {
		return nil, fmt.Errorf("update link: %w", err)
	}
	return &updatedLink, nil
}

func (repo *Repository) DeleteById(ctx context.Context, id int64) (int64, error) {
	findId, err := repo.queries.SoftDeleteLink(ctx, id)
	if err != nil {
		return findId, fmt.Errorf("delete link: %w", err)
	}
	if findId == 0 {
		return findId, ErrLinkNotFound
	}
	return findId, nil
}
