package link

import (
	"context"
	"fmt"
	"go-api/internal/database"
	"strings"
)

type LinkService interface {
	Create(
		ctx context.Context,
		rawURL string,
	) (*database.Link, error)
	GetByHash(
		ctx context.Context,
		hash string,
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

func (service *Service) Create(ctx context.Context, rawURL string) (*database.Link, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	link := NewLink(rawURL)

	createdLink, err := service.repository.Create(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	return createdLink, nil
}

func (service *Service) GetByHash(ctx context.Context, hash string) (*database.Link, error) {
	link, err := service.repository.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get link by hash: %w", err)
	}
	return link, nil
}
