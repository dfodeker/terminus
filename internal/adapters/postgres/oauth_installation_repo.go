// internal/adapters/postgres/oauth_installation_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInstallationNotFound = errors.New("oauth installation not found")
)

type OAuthInstallationRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewOAuthInstallationRepository(pool *pgxpool.Pool) *OAuthInstallationRepository {
	return &OAuthInstallationRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *OAuthInstallationRepository) Create(ctx context.Context, installation *domain.OAuthInstallation) error {
	row, err := r.queries.CreateOAuthInstallation(ctx, db.CreateOAuthInstallationParams{
		ShopID:          installation.ShopID,
		AppID:           installation.AppID,
		AccessTokenHash: installation.AccessTokenHash,
		Scopes:          scopesToStrings(installation.Scopes),
		Status:          string(installation.Status),
	})
	if err != nil {
		return err
	}

	installation.ID = row.ID
	installation.InstalledAt = row.InstalledAt
	installation.CreatedAt = row.CreatedAt
	installation.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *OAuthInstallationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OAuthInstallation, error) {
	row, err := r.queries.GetOAuthInstallationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstallationNotFound
		}
		return nil, err
	}

	return dbInstallationToDomain(row), nil
}

func (r *OAuthInstallationRepository) GetByAccessToken(ctx context.Context, tokenHash string) (*domain.OAuthInstallation, error) {
	row, err := r.queries.GetOAuthInstallationByAccessToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstallationNotFound
		}
		return nil, err
	}

	return dbInstallationToDomain(row), nil
}

func (r *OAuthInstallationRepository) Update(ctx context.Context, installation *domain.OAuthInstallation) error {
	row, err := r.queries.UpdateOAuthInstallation(ctx, db.UpdateOAuthInstallationParams{
		ID:              installation.ID,
		AccessTokenHash: textToPgtype(installation.AccessTokenHash),
		Scopes:          scopesToStrings(installation.Scopes),
		Status:          textToPgtype(string(installation.Status)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstallationNotFound
		}
		return err
	}

	installation.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *OAuthInstallationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteOAuthInstallation(ctx, id)
}

func (r *OAuthInstallationRepository) Uninstall(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.UninstallOAuthInstallation(ctx, id)
	return err
}

func dbInstallationToDomain(row db.OauthInstallation) *domain.OAuthInstallation {
	installation := &domain.OAuthInstallation{
		ID:              row.ID,
		ShopID:          row.ShopID,
		AppID:           row.AppID,
		AccessTokenHash: row.AccessTokenHash,
		Scopes:          stringsToScopes(row.Scopes),
		Status:          domain.InstallationStatus(row.Status),
		InstalledAt:     row.InstalledAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}

	if row.UninstalledAt.Valid {
		t := row.UninstalledAt.Time
		installation.UninstalledAt = &t
	}

	return installation
}
