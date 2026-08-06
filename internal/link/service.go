package link

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"
	"strings"
)

type LinkRepository interface {
	Create(
		ctx context.Context,
		link *database.Link,
	) (*database.Link, error)
}

type Service struct {
	repository LinkRepository
}

func NewLinkService(
	repository LinkRepository,
) *Service {
	return &Service{
		repository: repository,
	}
}

func (service *Service) Create(
	ctx context.Context,
	rawURL string,
) (*database.Link, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return nil, errors.New("URL is required")
	}

	link := NewLink(rawURL)

	createdLink, err := service.repository.Create(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	return createdLink, nil
}
