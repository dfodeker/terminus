// internal/adapters/postgres/product_repo.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProductRepository struct {
	tenantDB *TenantDB
	queries  *db.Queries
}

func NewProductRepository(tenantDB *TenantDB) *ProductRepository {
	return &ProductRepository{
		tenantDB: tenantDB,
		queries:  db.New(nil), // queries work with any DBTX
	}
}

func (r *ProductRepository) Create(ctx context.Context, shopID uuid.UUID, orgID uuid.UUID, product *domain.Product) error {
	return r.tenantDB.WithTenant(ctx, shopID, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, err := q.CreateProduct(ctx, db.CreateProductParams{
			ShopID:            shopID,
			OrganizationID:    orgID,
			Title:             product.Title,
			Description:       textToPgtype(product.Description),
			DescriptionHtml:   textToPgtype(product.BodyHTML),
			Handle:            product.Handle,
			PriceCents:        int64(product.PriceCents),
			Sku:               textToPgtype(product.SKU),
			InventoryQuantity: int32(product.Inventory),
			TrackInventory:    true,
			Status:            string(product.Status),
			// Optional fields with defaults
			CompareAtPriceCents: pgtype.Int8{Valid: false},
			CostCents:           pgtype.Int8{Valid: false},
			Vendor:              pgtype.Text{Valid: false},
			ProductType:         pgtype.Text{Valid: false},
			Tags:                []string{},
			Barcode:             pgtype.Text{Valid: false},
			SeoTitle:            pgtype.Text{Valid: false},
			SeoDescription:      pgtype.Text{Valid: false},
			TemplateSuffix:      pgtype.Text{Valid: false},
		})
		if err != nil {
			return fmt.Errorf("insert product: %w", err)
		}

		product.ID = domain.ProductID(row.ID)
		product.CreatedAt = row.CreatedAt
		product.UpdatedAt = row.UpdatedAt

		return nil
	})
}

func (r *ProductRepository) GetByID(ctx context.Context, shopID uuid.UUID, productID uuid.UUID) (*domain.Product, error) {
	var product *domain.Product

	err := r.tenantDB.WithTenant(ctx, shopID, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, err := q.GetProductByID(ctx, productID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return domain.ErrProductNotFound
			}
			return fmt.Errorf("get product: %w", err)
		}

		product = rowToProduct(row)
		return nil
	})

	return product, err
}

func (r *ProductRepository) List(ctx context.Context, shopID uuid.UUID, filter ports.ProductFilter) ([]domain.Product, string, error) {
	var products []domain.Product
	var nextCursor string

	err := r.tenantDB.WithTenant(ctx, shopID, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		// Parse cursor if provided
		var cursor pgtype.Timestamptz
		if filter.Cursor != "" {
			t, err := time.Parse(time.RFC3339Nano, filter.Cursor)
			if err == nil {
				cursor = pgtype.Timestamptz{Time: t, Valid: true}
			}
		}

		limit := int32(filter.Limit)
		if limit == 0 {
			limit = 20
		}

		rows, err := q.ListProductsByShop(ctx, db.ListProductsByShopParams{
			ShopID: shopID,
			Status: statusToPgtype(filter.Status),
			Cursor: cursor,
			Limit:  limit + 1, // Fetch one extra for cursor
		})
		if err != nil {
			return fmt.Errorf("list products: %w", err)
		}

		for i, row := range rows {
			if int32(i) >= limit {
				// This is the extra row - use it for cursor
				nextCursor = row.CreatedAt.Format(time.RFC3339Nano)
				break
			}
			products = append(products, *rowToProduct(row))
		}

		return nil
	})

	return products, nextCursor, err
}

func (r *ProductRepository) DeductInventory(ctx context.Context, shopID uuid.UUID, productID uuid.UUID, qty int) error {
	return r.tenantDB.WithTenant(ctx, shopID, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		_, err := q.DeductInventory(ctx, db.DeductInventoryParams{
			ID:       productID,
			Quantity: int32(qty),
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return domain.ErrInsufficientInventory
			}
			return fmt.Errorf("deduct inventory: %w", err)
		}

		return nil
	})
}

// Helper functions
func rowToProduct(row db.Product) *domain.Product {
	return &domain.Product{
		ID:          domain.ProductID(row.ID),
		ShopID:      domain.ShopID(row.ShopID),
		Title:       row.Title,
		Description: pgtypeToString(row.Description),
		BodyHTML:    pgtypeToString(row.DescriptionHtml),
		Handle:      row.Handle,
		PriceCents:  domain.Money(row.PriceCents),
		SKU:         pgtypeToString(row.Sku),
		Inventory:   int(row.InventoryQuantity),
		Status:      domain.ProductStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// pgtypeToString extracts string from pgtype.Text, returning empty string if not valid
func pgtypeToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}


// statusToPgtype converts a product status pointer to pgtype.Text
func statusToPgtype(s *domain.ProductStatus) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: string(*s), Valid: true}
}
