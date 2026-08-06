package link

import (
	"context"
	"fmt"
	"go-api/internal/database"
)

type Repository struct {
	queries database.Querier
}

func NewLinkRepository(queries database.Querier) *Repository {
	return &Repository{
		queries: queries,
	}
}

func (repo *Repository) Create(
	ctx context.Context,
	link *database.Link,
) (*database.Link, error) {
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
