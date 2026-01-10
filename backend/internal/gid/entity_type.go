package gid

// EntityType represents the type of entity a GID refers to
type EntityType string

const (
	EntityProduct        EntityType = "Product"
	EntityProductVariant EntityType = "ProductVariant"
	EntityStore          EntityType = "Store"
	EntityTenant         EntityType = "Tenant"
	EntityUser           EntityType = "User"
	EntityRole           EntityType = "Role"
	EntityPermission     EntityType = "Permission"
	EntityCustomDomain   EntityType = "CustomDomain"
)

// ValidEntityTypes maps valid entity types for validation
var ValidEntityTypes = map[EntityType]bool{
	EntityProduct:        true,
	EntityProductVariant: true,
	EntityStore:          true,
	EntityTenant:         true,
	EntityUser:           true,
	EntityRole:           true,
	EntityPermission:     true,
	EntityCustomDomain:   true,
}

// IsValid checks if the entity type is valid
func (e EntityType) IsValid() bool {
	return ValidEntityTypes[e]
}
