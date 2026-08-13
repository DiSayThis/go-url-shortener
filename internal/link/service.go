package link

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type LinkService interface {
	Create(ctx context.Context, userID int64, rawURL string) (*database.Link, error)
	GetByHash(ctx context.Context, hash string) (*database.Link, error)
	Update(ctx context.Context, userID int64, publicID, url, hash string) (*database.Link, error)
	Delete(ctx context.Context, userID int64, publicID string) error
}

type Service struct {
	repository LinkRepository
}

func NewLinkService(repository LinkRepository) *Service {
	return &Service{repository: repository}
}

const (
	linkHashLength    = 22
	maxCreateAttempts = 5
)

func (service *Service) Create(ctx context.Context, userID int64, rawURL string) (*database.Link, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	for range maxCreateAttempts {
		hash, err := GenerateHash(linkHashLength)
		if err != nil {
			return nil, fmt.Errorf("generate link hash: %w", err)
		}

		createdLink, err := service.repository.Create(ctx, userID, NewLink(rawURL, hash))
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
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return nil, ErrLinkNotFound
	}

	foundLink, err := service.repository.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get link by hash: %w", err)
	}

	return foundLink, nil
}

func (service *Service) Update(
	ctx context.Context,
	userID int64,
	publicID string,
	url string,
	hash string,
) (*database.Link, error) {
	parsedPublicID, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return nil, ErrInvalidURL
	}

	hash = strings.TrimSpace(hash)
	if hash != "" {
		updatedLink, err := service.repository.UpdateWithHash(
			ctx,
			userID,
			parsedPublicID,
			url,
			hash,
		)
		if err != nil {
			return nil, fmt.Errorf("update link: %w", err)
		}
		return updatedLink, nil
	}

	updatedLink, err := service.repository.Update(ctx, userID, parsedPublicID, url)
	if err != nil {
		return nil, fmt.Errorf("update link: %w", err)
	}

	return updatedLink, nil
}

func (service *Service) Delete(ctx context.Context, userID int64, publicID string) error {
	parsedPublicID, err := parsePublicID(publicID)
	if err != nil {
		return err
	}

	if err := service.repository.Delete(ctx, userID, parsedPublicID); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}

	return nil
}

func parsePublicID(value string) (pgtype.UUID, error) {
	var publicID pgtype.UUID
	if err := publicID.Scan(strings.TrimSpace(value)); err != nil || !publicID.Valid {
		return pgtype.UUID{}, ErrInvalidLinkID
	}

	return publicID, nil
}
