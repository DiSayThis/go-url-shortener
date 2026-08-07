package link

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"
	"strings"
)

type LinkService interface {
	Create(ctx context.Context, rawURL string) (*database.Link, error)
	GetByHash(ctx context.Context, hash string) (*database.Link, error)
}

type Service struct {
	repository LinkRepository
}

func NewLinkService(repository LinkRepository) *Service {
	return &Service{
		repository: repository,
	}
}

const (
	linkHashLength    = 22
	maxCreateAttempts = 5
)

func (service *Service) Create(ctx context.Context, rawURL string) (*database.Link, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	for range maxCreateAttempts {
		hash, err := GenerateHash(linkHashLength)
		if err != nil {
			return nil, fmt.Errorf("generate link hash: %w", err)
		}

		link := NewLink(rawURL, hash)

		createdLink, err := service.repository.Create(ctx, link)
		if err == nil {
			return createdLink, nil
		}

		if !errors.Is(err, ErrHashCollision) {
			return nil, fmt.Errorf("create link: %w", err)
		}
	}

	return nil, fmt.Errorf(
		"create link after %d attempts: %w",
		maxCreateAttempts,
		ErrHashCollision,
	)
}

func (service *Service) GetByHash(ctx context.Context, hash string) (*database.Link, error) {
	link, err := service.repository.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get link by hash: %w", err)
	}
	return link, nil
}
