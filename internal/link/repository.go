package link

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"

	"github.com/jackc/pgx/v5"
)

type LinkRepository interface {
	Create(
		ctx context.Context,
		link *database.Link,
	) (*database.Link, error)
	GetByHash(
		ctx context.Context,
		hash string,
	) (*database.Link, error)
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
