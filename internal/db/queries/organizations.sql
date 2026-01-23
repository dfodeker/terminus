-- db/queries/organizations.sql

-- name: CreateOrganization :one
INSERT INTO organizations (
    name, handle, plan, billing_email, stripe_customer_id,
    status, max_shops, max_members, settings
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrganizationByHandle :one
SELECT * FROM organizations WHERE handle = $1 AND deleted_at IS NULL;

-- name: GetOrganizationByStripeCustomer :one
SELECT * FROM organizations WHERE stripe_customer_id = $1 AND deleted_at IS NULL;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = COALESCE(sqlc.narg('name'), name),
    handle = COALESCE(sqlc.narg('handle'), handle),
    billing_email = COALESCE(sqlc.narg('billing_email'), billing_email),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateOrganizationPlan :one
UPDATE organizations
SET 
    plan = $2,
    max_shops = $3,
    max_members = $4,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateOrganizationStatus :one
UPDATE organizations
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SetStripeCustomerID :one
UPDATE organizations
SET stripe_customer_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrganization :exec
UPDATE organizations SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations
WHERE deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: CountOrganizations :one
SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL;

-- name: OrganizationExistsByHandle :one
SELECT EXISTS(SELECT 1 FROM organizations WHERE handle = $1 AND deleted_at IS NULL);

-- =============================================================================
-- ORGANIZATION MEMBERS
-- =============================================================================

-- name: AddOrganizationMember :one
INSERT INTO organization_members (
    organization_id, user_id, role, invited_by, invited_at
)
VALUES ($1, $2, $3, $4, NOW())
RETURNING *;

-- name: GetOrganizationMember :one
SELECT * FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: ListOrganizationMembers :many
SELECT 
    om.*,
    u.email,
    u.first_name,
    u.last_name,
    u.avatar_url
FROM organization_members om
JOIN users u ON u.id = om.user_id
WHERE om.organization_id = $1
ORDER BY om.created_at DESC;

-- name: ListUserOrganizations :many
SELECT 
    o.*,
    om.role
FROM organizations o
JOIN organization_members om ON om.organization_id = o.id
WHERE om.user_id = $1 AND o.deleted_at IS NULL
ORDER BY o.created_at DESC;

-- name: UpdateOrganizationMemberRole :one
UPDATE organization_members
SET role = $3, updated_at = NOW()
WHERE organization_id = $1 AND user_id = $2
RETURNING *;

-- name: AcceptOrganizationInvite :one
UPDATE organization_members
SET accepted_at = NOW(), updated_at = NOW()
WHERE organization_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: CountOrganizationMembers :one
SELECT COUNT(*) FROM organization_members WHERE organization_id = $1;

-- name: IsOrganizationOwner :one
SELECT EXISTS(
    SELECT 1 FROM organization_members 
    WHERE organization_id = $1 AND user_id = $2 AND role = 'owner'
);

-- name: IsOrganizationAdmin :one
SELECT EXISTS(
    SELECT 1 FROM organization_members 
    WHERE organization_id = $1 AND user_id = $2 AND role IN ('owner', 'admin')
);
