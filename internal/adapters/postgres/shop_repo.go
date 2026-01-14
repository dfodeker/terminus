// internal/adapters/postgres/shop_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrShopNotFound      = errors.New("shop not found")
	ErrShopAlreadyExists = errors.New("shop with this subdomain already exists")
)

type ShopRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewShopRepository(pool *pgxpool.Pool) *ShopRepository {
	return &ShopRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *ShopRepository) Create(ctx context.Context, shop *domain.Shop) error {
	row, err := r.queries.CreateShop(ctx, db.CreateShopParams{
		OrganizationID: shop.OrganizationID,
		Name:           shop.Name,
		Handle:         shop.MyShopifyDomain, // using subdomain as handle for now
		Subdomain:      shop.MyShopifyDomain,
		CustomDomain:   textToPgtype(shop.Domain),
		Currency:       shop.Currency,
		Locale:         "en", // default locale
		Timezone:       shop.Timezone,
		ShopOwner:      textToPgtype(shop.ShopOwner),
		Email:          textToPgtype(shop.CustomerEmail),
		Phone:          textToPgtype(shop.Phone),
		Source:         shop.Source,
		ReferralCode:   pgtype.Text{Valid: false},
		Status:         "active",
		Gid:            pgtype.Int8{Int64: int64(shop.Id), Valid: shop.Id != 0},
	})
	if err != nil {
		// Check for unique constraint violation
		if isPgUniqueViolation(err) {
			return ErrShopAlreadyExists
		}
		return fmt.Errorf("create shop: %w", err)
	}

	// Map returned row back to domain
	if row.Gid.Valid {
		shop.Id = uint64(row.Gid.Int64)
	}
	shop.CreatedAt = &row.CreatedAt
	shop.UpdatedAt = &row.UpdatedAt

	return nil
}

func (r *ShopRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shop, error) {
	row, err := r.queries.GetShopByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("get shop by id: %w", err)
	}

	return rowToShop(row), nil
}

func (r *ShopRepository) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Shop, error) {
	row, err := r.queries.GetShopBySubdomain(ctx, subdomain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("get shop by subdomain: %w", err)
	}

	return rowToShop(row), nil
}

func (r *ShopRepository) GetByCustomDomain(ctx context.Context, customDomain string) (*domain.Shop, error) {
	row, err := r.queries.GetShopByCustomDomain(ctx, pgtype.Text{String: customDomain, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("get shop by custom domain: %w", err)
	}

	return rowToShop(row), nil
}

func (r *ShopRepository) Update(ctx context.Context, shop *domain.Shop) error {
	// Convert uint64 ID to UUID - this assumes you have a mapping
	// You may need to adjust based on your actual ID strategy
	id, err := r.getUUIDFromShopID(ctx, shop.Id)
	if err != nil {
		return fmt.Errorf("get uuid for shop: %w", err)
	}

	row, err := r.queries.UpdateShop(ctx, db.UpdateShopParams{
		ID:           id,
		Name:         textToPgtype(shop.Name),
		Handle:       pgtype.Text{Valid: false}, // don't update handle
		CustomDomain: textToPgtype(shop.Domain),
		Currency:     textToPgtype(shop.Currency),
		Locale:       pgtype.Text{Valid: false}, // don't update locale
		Timezone:     textToPgtype(shop.Timezone),
		ShopOwner:    textToPgtype(shop.ShopOwner),
		Email:        textToPgtype(shop.CustomerEmail),
		Phone:        textToPgtype(shop.Phone),
		Status:       pgtype.Text{Valid: false}, // don't update status
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrShopNotFound
		}
		return fmt.Errorf("update shop: %w", err)
	}

	shop.UpdatedAt = &row.UpdatedAt

	return nil
}

func (r *ShopRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete - set deleted_at timestamp
	_, err := r.pool.Exec(ctx,
		"UPDATE shops SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return fmt.Errorf("delete shop: %w", err)
	}
	return nil
}

func (r *ShopRepository) List(ctx context.Context, filter ports.ShopFilter) ([]domain.Shop, string, error) {
	limit := int32(filter.Limit)
	if limit == 0 {
		limit = 20
	}

	// Parse cursor if provided
	var cursor *time.Time
	if filter.Cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, filter.Cursor)
		if err == nil {
			cursor = &t
		}
	}

	// Build query with optional filters
	query := `
		SELECT id, name, subdomain, custom_domain, currency, timezone, 
		       shop_owner, phone, plan_name, source, referral_code, 
		       status, gid, created_at, updated_at
		FROM shops 
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argIndex := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if cursor != nil {
		query += fmt.Sprintf(" AND created_at < $%d", argIndex)
		args = append(args, *cursor)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIndex)
	args = append(args, limit+1) // Fetch one extra to determine if there's more

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list shops: %w", err)
	}
	defer rows.Close()

	var shops []domain.Shop
	for rows.Next() {
		var row shopRow
		err := rows.Scan(
			&row.ID, &row.Name, &row.Subdomain, &row.CustomDomain,
			&row.Currency, &row.Timezone, &row.ShopOwner, &row.Phone,
			&row.PlanName, &row.Source, &row.ReferralCode, &row.Status,
			&row.GID, &row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("scan shop row: %w", err)
		}
		shops = append(shops, *shopRowToDomain(&row))
	}

	var nextCursor string
	if len(shops) > int(limit) {
		shops = shops[:limit]
		lastShop := shops[len(shops)-1]
		if lastShop.CreatedAt != nil {
			nextCursor = lastShop.CreatedAt.Format(time.RFC3339Nano)
		}
	}

	return shops, nextCursor, nil
}

// Helper to get UUID from snowflake ID
func (r *ShopRepository) getUUIDFromShopID(ctx context.Context, gid uint64) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT id FROM shops WHERE gid = $1", gid).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// shopRow is an intermediate type for scanning
type shopRow struct {
	ID           uuid.UUID
	Name         string
	Subdomain    string
	CustomDomain *string
	Currency     string
	Timezone     string
	ShopOwner    *string
	Phone        *string
	PlanName     string
	Source       string
	ReferralCode *string
	Status       string
	GID          *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func shopRowToDomain(row *shopRow) *domain.Shop {
	shop := &domain.Shop{
		Name:            row.Name,
		MyShopifyDomain: row.Subdomain,
		Currency:        row.Currency,
		Timezone:        row.Timezone,
		PlanName:        row.PlanName,
		Source:          row.Source,
		CreatedAt:       &row.CreatedAt,
		UpdatedAt:       &row.UpdatedAt,
	}

	if row.GID != nil {
		shop.Id = uint64(*row.GID)
	}
	if row.CustomDomain != nil {
		shop.Domain = *row.CustomDomain
	}
	if row.ShopOwner != nil {
		shop.ShopOwner = *row.ShopOwner
	}
	if row.Phone != nil {
		shop.Phone = *row.Phone
	}

	return shop
}

// rowToShop converts a sqlc-generated row to domain.Shop
func rowToShop(row db.Shop) *domain.Shop {
	shop := &domain.Shop{
		Name:            row.Name,
		MyShopifyDomain: row.Subdomain,
		Currency:        row.Currency,
		Timezone:        row.Timezone,
		Source:          row.Source,
		CreatedAt:       &row.CreatedAt,
		UpdatedAt:       &row.UpdatedAt,
	}

	if row.Gid.Valid {
		shop.Id = uint64(row.Gid.Int64)
	}
	if row.CustomDomain.Valid {
		shop.Domain = row.CustomDomain.String
	}
	if row.ShopOwner.Valid {
		shop.ShopOwner = row.ShopOwner.String
	}
	if row.Phone.Valid {
		shop.Phone = row.Phone.String
	}
	if row.Email.Valid {
		shop.CustomerEmail = row.Email.String
	}

	return shop
}

