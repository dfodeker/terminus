package organization

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	repo     ports.OrganizationRepository
	shopRepo ports.ShopRepository
	cache    ports.Cache
	events   ports.EventPublisher
}

func NewService(
	repo ports.OrganizationRepository,
	shopRepo ports.ShopRepository,
	cache ports.Cache,
	events ports.EventPublisher,
) *Service {
	return &Service{
		repo:     repo,
		shopRepo: shopRepo,
		cache:    cache,
		events:   events,
	}
}

type CreateOrganizationCmd struct {
	Name         string
	Handle       string
	BillingEmail string
	Plan         string
}

type UpdateOrganizationCmd struct {
	ID               uuid.UUID
	Name             *string
	Handle           *string
	BillingEmail     *string
	Plan             *string
	StripeCustomerID *string
	Status           *string
	MaxShops         *int32
	MaxMembers       *int32
	Settings         json.RawMessage
}

// Create creates a new organization
func (s *Service) Create(ctx context.Context, cmd CreateOrganizationCmd) (*domain.Organization, error) {
	plan := cmd.Plan
	if plan == "" {
		plan = string(domain.OrganizationPlanFree)
	}

	// Set default limits based on plan
	maxShops, maxMembers := s.getPlanLimits(plan)

	org := &domain.Organization{
		Name:         cmd.Name,
		Handle:       cmd.Handle,
		BillingEmail: cmd.BillingEmail,
		Plan:         plan,
		Status:       "active",
		MaxShops:     maxShops,
		MaxMembers:   maxMembers,
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	// Publish event
	s.events.Publish(ctx, OrganizationCreatedEvent{Organization: org})

	return org, nil
}

// GetByID retrieves an organization by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	cacheKey := fmt.Sprintf("org:%s", id.String())
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if org, ok := cached.(*domain.Organization); ok {
			return org, nil
		}
	}

	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, org, 0)

	return org, nil
}

// GetByHandle retrieves an organization by handle
func (s *Service) GetByHandle(ctx context.Context, handle string) (*domain.Organization, error) {
	cacheKey := fmt.Sprintf("org:handle:%s", handle)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if org, ok := cached.(*domain.Organization); ok {
			return org, nil
		}
	}

	org, err := s.repo.GetByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, org, 0)

	return org, nil
}

// Update updates an organization
func (s *Service) Update(ctx context.Context, cmd UpdateOrganizationCmd) (*domain.Organization, error) {
	org, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if cmd.Name != nil {
		org.Name = *cmd.Name
	}
	if cmd.Handle != nil {
		org.Handle = *cmd.Handle
	}
	if cmd.BillingEmail != nil {
		org.BillingEmail = *cmd.BillingEmail
	}
	if cmd.Plan != nil {
		org.Plan = *cmd.Plan
	}
	if cmd.StripeCustomerID != nil {
		org.StripeCustomerID = *cmd.StripeCustomerID
	}
	if cmd.Status != nil {
		org.Status = *cmd.Status
	}
	if cmd.MaxShops != nil {
		org.MaxShops = *cmd.MaxShops
	}
	if cmd.MaxMembers != nil {
		org.MaxMembers = *cmd.MaxMembers
	}
	if cmd.Settings != nil {
		org.Settings = cmd.Settings
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("update organization: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, org)

	// Publish event
	s.events.Publish(ctx, OrganizationUpdatedEvent{Organization: org})

	return org, nil
}

// Delete deletes an organization
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if organization has shops
	shopCount, err := s.repo.CountShops(ctx, id)
	if err != nil {
		return fmt.Errorf("count shops: %w", err)
	}
	if shopCount > 0 {
		return fmt.Errorf("cannot delete organization with %d active shops", shopCount)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, org)

	// Publish event
	s.events.Publish(ctx, OrganizationDeletedEvent{OrganizationID: id})

	return nil
}

// List returns a paginated list of organizations
func (s *Service) List(ctx context.Context, filter ports.OrganizationFilter) ([]domain.Organization, string, error) {
	return s.repo.List(ctx, filter)
}

// ListShops returns shops belonging to an organization
func (s *Service) ListShops(ctx context.Context, orgID uuid.UUID, filter ports.ShopFilter) ([]domain.Shop, string, error) {
	filter.OrganizationID = &orgID
	return s.shopRepo.List(ctx, filter)
}

// CanCreateShop checks if the organization can create another shop
// Returns: allowed, currentCount, maxAllowed, error
func (s *Service) CanCreateShop(ctx context.Context, orgID uuid.UUID) (bool, int, int32, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return false, 0, 0, err
	}

	shopCount, err := s.repo.CountShops(ctx, orgID)
	if err != nil {
		return false, 0, 0, fmt.Errorf("count shops: %w", err)
	}

	allowed := org.CanCreateStore(shopCount) == nil
	return allowed, shopCount, org.MaxShops, nil
}

// CanAddMember checks if the organization can add another member
func (s *Service) CanAddMember(ctx context.Context, orgID uuid.UUID) error {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return err
	}

	memberCount, err := s.repo.CountMembers(ctx, orgID)
	if err != nil {
		return fmt.Errorf("count members: %w", err)
	}

	return org.CanAddMembers(memberCount)
}

// UpgradePlan upgrades an organization's plan
func (s *Service) UpgradePlan(ctx context.Context, orgID uuid.UUID, newPlan string) (*domain.Organization, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.Plan = newPlan
	maxShops, maxMembers := s.getPlanLimits(newPlan)
	org.MaxShops = maxShops
	org.MaxMembers = maxMembers

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("update organization plan: %w", err)
	}

	s.invalidateCache(ctx, org)

	s.events.Publish(ctx, OrganizationPlanChangedEvent{
		Organization: org,
		OldPlan:      org.Plan,
		NewPlan:      newPlan,
	})

	return org, nil
}

func (s *Service) getPlanLimits(plan string) (maxShops int32, maxMembers int32) {
	switch domain.OrganizationPlan(plan) {
	case domain.OrganizationPlanFree:
		return 1, 2
	case domain.OrganizationPlanPro:
		return 10, 25
	default:
		return 1, 2
	}
}

func (s *Service) invalidateCache(ctx context.Context, org *domain.Organization) {
	s.cache.Delete(ctx, fmt.Sprintf("org:%s", uuid.UUID(org.ID).String()))
	if org.Handle != "" {
		s.cache.Delete(ctx, fmt.Sprintf("org:handle:%s", org.Handle))
	}
}

// Events

type OrganizationCreatedEvent struct {
	Organization *domain.Organization
}

type OrganizationUpdatedEvent struct {
	Organization *domain.Organization
}

type OrganizationDeletedEvent struct {
	OrganizationID uuid.UUID
}

type OrganizationPlanChangedEvent struct {
	Organization *domain.Organization
	OldPlan      string
	NewPlan      string
}
