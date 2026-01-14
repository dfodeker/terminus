package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type OrganizationID uuid.UUID
type OrganizationPlan string

type Organization struct {
	ID               OrganizationID
	Name             string          `json:"name"`
	Handle           string          `json:"handle"`
	Plan             string          `json:"plan"`
	BillingEmail     string          `json:"billing_email"`
	StripeCustomerID string          `json:"stripe_customer_id"`
	Status           string          `json:"status"`
	MaxShops         int32           `json:"max_shops"`
	MaxMembers       int32           `json:"max_members"`
	Settings         json.RawMessage `json:"settings"`
}

const (
	OrganizationPlanFree OrganizationPlan = "free"
	OrganizationPlanPro  OrganizationPlan = "pro"
)

func (o *Organization) CanCreateStore(storeCount int) error {
	if storeCount+1 > int(o.MaxShops) {
		return ErrShopCreateDisable
	}
	return nil
}

func (o *Organization) CanAddMembers(memberCount int) error {
	if memberCount+1 > int(o.MaxMembers) {
		return ErrShopCreateDisable //needs to be changed
	}
	return nil
}
