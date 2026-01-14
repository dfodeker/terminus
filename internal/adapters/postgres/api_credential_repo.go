// internal/adapters/postgres/api_credential_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCredentialNotFound = errors.New("api credential not found")
)

type APICredentialRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewAPICredentialRepository(pool *pgxpool.Pool) *APICredentialRepository {
	return &APICredentialRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *APICredentialRepository) Create(ctx context.Context, cred *domain.APICredential) error {
	row, err := r.queries.CreateAPICredential(ctx, db.CreateAPICredentialParams{
		OrganizationID: uuid.Nil, // TODO: Add organization ID to domain model
		ShopID:         uuidToPgtype(cred.ShopID),
		Name:           cred.Name,
		KeyPrefix:      cred.KeyPrefix,
		KeyHash:        cred.KeyHash,
		Scopes:         scopesToStrings(cred.Scopes),
		Environment:    "live", // TODO: Add environment to domain model
		Status:         string(cred.Status),
		CreatedBy:      pgtype.UUID{Valid: false},
		ExpiresAt:      timePtrToPgtype(cred.ExpiresAt),
	})
	if err != nil {
		return err
	}

	cred.ID = row.ID
	cred.CreatedAt = row.CreatedAt

	return nil
}

func (r *APICredentialRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APICredential, error) {
	row, err := r.queries.GetAPICredentialByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	return dbCredentialToDomain(row), nil
}

func (r *APICredentialRepository) GetByPrefix(ctx context.Context, prefix string) (*domain.APICredential, error) {
	row, err := r.queries.GetAPICredentialByPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	return dbCredentialToDomain(row), nil
}

func (r *APICredentialRepository) Update(ctx context.Context, cred *domain.APICredential) error {
	_, err := r.queries.UpdateAPICredential(ctx, db.UpdateAPICredentialParams{
		ID:        cred.ID,
		Name:      textToPgtype(cred.Name),
		Scopes:    scopesToStrings(cred.Scopes),
		ExpiresAt: timePtrToPgtype(cred.ExpiresAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCredentialNotFound
		}
		return err
	}

	return nil
}

func (r *APICredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteAPICredential(ctx, id)
}

func (r *APICredentialRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateAPICredentialLastUsed(ctx, id)
}

func (r *APICredentialRepository) ListByShop(ctx context.Context, shopID uuid.UUID) ([]domain.APICredential, error) {
	rows, err := r.queries.ListAPICredentialsByShop(ctx, db.ListAPICredentialsByShopParams{
		ShopID: uuidToPgtype(shopID),
		Status: pgtype.Text{Valid: false}, // All statuses
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}

	creds := make([]domain.APICredential, len(rows))
	for i, row := range rows {
		creds[i] = *dbCredentialToDomain(row)
	}

	return creds, nil
}

func (r *APICredentialRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.RevokeAPICredential(ctx, id)
	return err
}

func dbCredentialToDomain(row db.ApiCredential) *domain.APICredential {
	cred := &domain.APICredential{
		ID:           row.ID,
		Name:         row.Name,
		KeyPrefix:    row.KeyPrefix,
		KeyHash:      row.KeyHash,
		Scopes:       stringsToScopes(row.Scopes),
		Status:       domain.CredentialStatus(row.Status),
		RequestCount: 0, // Not stored in DB currently
		CreatedAt:    row.CreatedAt,
	}

	if row.ShopID.Valid {
		cred.ShopID = row.ShopID.Bytes
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		cred.LastUsedAt = &t
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		cred.ExpiresAt = &t
	}

	return cred
}

func scopesToStrings(scopes []domain.Scope) []string {
	result := make([]string, len(scopes))
	for i, s := range scopes {
		result[i] = string(s)
	}
	return result
}

func stringsToScopes(strs []string) []domain.Scope {
	result := make([]domain.Scope, len(strs))
	for i, s := range strs {
		result[i] = domain.Scope(s)
	}
	return result
}
