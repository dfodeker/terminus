// internal/domain/scopes.go
package domain

type Scope string

const (
	// Read scopes
	ScopeProductsRead      Scope = "read_products"
	ScopeOrdersRead        Scope = "read_orders"
	ScopeCustomersRead     Scope = "read_customers"
	ScopeInventoryRead     Scope = "read_inventory"
	ScopeShopsRead         Scope = "read_shops"
	ScopeOrganizationsRead Scope = "read_organizations"

	// Write scopes
	ScopeProductsWrite      Scope = "write_products"
	ScopeOrdersWrite        Scope = "write_orders"
	ScopeCustomersWrite     Scope = "write_customers"
	ScopeInventoryWrite     Scope = "write_inventory"
	ScopeShopsWrite         Scope = "write_shops"
	ScopeOrganizationsWrite Scope = "write_organizations"

	// Special scopes
	ScopeWebhooksManage Scope = "manage_webhooks"
	ScopeScriptTags     Scope = "write_script_tags"
)

// ScopeDefinitions describes what each scope grants
var ScopeDefinitions = map[Scope]ScopeDefinition{
	ScopeProductsRead: {
		Name:        "Read Products",
		Description: "View products, variants, and collections",
		Resources:   []string{"products", "variants", "collections"},
		Actions:     []string{"read"},
	},
	ScopeProductsWrite: {
		Name:        "Write Products",
		Description: "Create, update, and delete products",
		Resources:   []string{"products", "variants", "collections"},
		Actions:     []string{"read", "create", "update", "delete"},
		Implies:     []Scope{ScopeProductsRead},
	},
	// ... etc
}

type ScopeDefinition struct {
	Name        string
	Description string
	Resources   []string
	Actions     []string
	Implies     []Scope // Write implies Read
}

// ScopeSet for efficient checking
type ScopeSet map[Scope]struct{}

func NewScopeSet(scopes []Scope) ScopeSet {
	set := make(ScopeSet)
	for _, s := range scopes {
		set[s] = struct{}{}
		// Add implied scopes
		if def, ok := ScopeDefinitions[s]; ok {
			for _, implied := range def.Implies {
				set[implied] = struct{}{}
			}
		}
	}
	return set
}

func (s ScopeSet) Has(scope Scope) bool {
	_, ok := s[scope]
	return ok
}

func (s ScopeSet) HasAny(scopes ...Scope) bool {
	for _, scope := range scopes {
		if s.Has(scope) {
			return true
		}
	}
	return false
}
