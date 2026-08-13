package link

import (
	"context"
	"errors"
	"fmt"
	"go-api/internal/database"
	"strconv"
	"strings"
)

type LinkService interface {
	Create(ctx context.Context, rawURL string) (*database.Link, error)
	GetByHash(ctx context.Context, hash string) (*database.Link, error)
	UpdateLinkAndHashById(ctx context.Context, id, url, hash string) (*database.Link, error)
	DeleteById(ctx context.Context, id string) (int64, error)
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

func (service *Service) UpdateLinkAndHashById(ctx context.Context, id, url, hash string) (*database.Link, error) {
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return nil, ErrInvalidURL
	}

	if hash != "" {
		updatedLink, err := service.repository.UpdateWithHash(ctx, &database.Link{ID: idInt64, Url: url, Hash: hash})
		if err != nil {
			return nil, fmt.Errorf("update link: %w", err)
		}
		return updatedLink, nil
	}

	updatedLink, err := service.repository.Update(ctx, &database.Link{ID: idInt64, Url: url})
	if err != nil {
		return nil, fmt.Errorf("update link: %w", err)
	}
	return updatedLink, nil

}

func (service *Service) DeleteById(ctx context.Context, id string) (int64, error) {
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %w", err)
	}
	findId, err := service.repository.DeleteById(ctx, idInt64)
	if err != nil {
		return findId, fmt.Errorf("delete link by id: %w", err)
	}
	return findId, nil
}
