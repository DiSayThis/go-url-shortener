package link

import (
	"context"
	"errors"
	"testing"

	"go-api/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type linkRepositoryStub struct {
	create func(context.Context, int64, *database.Link) (*database.Link, error)
	update func(context.Context, int64, pgtype.UUID, string) (*database.Link, error)
	delete func(context.Context, int64, pgtype.UUID) error
}

func (stub linkRepositoryStub) Create(
	ctx context.Context,
	userID int64,
	link *database.Link,
) (*database.Link, error) {
	return stub.create(ctx, userID, link)
}

func (stub linkRepositoryStub) GetByHash(context.Context, string) (*database.Link, error) {
	return nil, ErrLinkNotFound
}

func (stub linkRepositoryStub) Update(
	ctx context.Context,
	userID int64,
	publicID pgtype.UUID,
	url string,
) (*database.Link, error) {
	if stub.update == nil {
		return nil, errors.New("unexpected update call")
	}
	return stub.update(ctx, userID, publicID, url)
}

func (stub linkRepositoryStub) UpdateWithHash(
	context.Context,
	int64,
	pgtype.UUID,
	string,
	string,
) (*database.Link, error) {
	return nil, errors.New("unexpected update with hash call")
}

func (stub linkRepositoryStub) Delete(
	ctx context.Context,
	userID int64,
	publicID pgtype.UUID,
) error {
	if stub.delete == nil {
		return errors.New("unexpected delete call")
	}
	return stub.delete(ctx, userID, publicID)
}

func TestServiceCreatePassesOwnerToRepository(t *testing.T) {
	const userID int64 = 42

	repository := linkRepositoryStub{
		create: func(_ context.Context, gotUserID int64, link *database.Link) (*database.Link, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %d, want %d", gotUserID, userID)
			}
			if link.Url != "https://example.com" {
				t.Fatalf("URL = %q, want %q", link.Url, "https://example.com")
			}
			if len(link.Hash) != linkHashLength {
				t.Fatalf("hash length = %d, want %d", len(link.Hash), linkHashLength)
			}

			link.UserID = gotUserID
			return link, nil
		},
	}

	service := NewLinkService(repository)
	createdLink, err := service.Create(context.Background(), userID, " https://example.com ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if createdLink.UserID != userID {
		t.Errorf("created link userID = %d, want %d", createdLink.UserID, userID)
	}
}

func TestServiceUpdateRejectsInvalidPublicID(t *testing.T) {
	service := NewLinkService(linkRepositoryStub{})

	_, err := service.Update(
		context.Background(),
		42,
		"not-a-uuid",
		"https://example.com",
		"",
	)
	if !errors.Is(err, ErrInvalidLinkID) {
		t.Fatalf("Update() error = %v, want ErrInvalidLinkID", err)
	}
}

func TestServiceDeletePassesOwnerToRepository(t *testing.T) {
	const (
		userID   int64 = 42
		publicID       = "0198a8f2-fcd7-7c12-9f78-8c14f5d3f001"
	)

	repository := linkRepositoryStub{
		delete: func(_ context.Context, gotUserID int64, gotPublicID pgtype.UUID) error {
			if gotUserID != userID {
				t.Fatalf("userID = %d, want %d", gotUserID, userID)
			}
			if formatUUID(gotPublicID) != publicID {
				t.Fatalf("publicID = %q, want %q", formatUUID(gotPublicID), publicID)
			}
			return nil
		},
	}

	service := NewLinkService(repository)
	if err := service.Delete(context.Background(), userID, publicID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
