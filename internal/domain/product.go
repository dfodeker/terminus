package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductID uuid.UUID
type ShopID uuid.UUID
type Money int64
type ProductStatus string
type Product struct {
	ID          ProductID
	ShopID      ShopID
	Title       string
	Description string
	Handle      string
	BodyHTML    string
	PriceCents  Money
	SKU         string
	Inventory   int //int for now will migrate to a seperate repo so we can have locations
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusArchived ProductStatus = "archived"
)

func (p *Product) CanPurchase(qty int) error {
	if p.Status != ProductStatusActive {
		return ErrProductNotAvailable
	}
	if p.Inventory < qty {
		return ErrInsufficientInventory
	}
	return nil
}

// ProductCreatedEvent is published when a new product is created.
type ProductCreatedEvent struct {
	Product *Product
}
