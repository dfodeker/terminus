package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/internal/gid"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

const ownerRoleName = "Owner"
const ownerRoleDesc = "Full access to all tenant resources, stores, and settings"

type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	GID       string    `json:"gid"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantUserResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTenantResponse struct {
	Tenant     TenantResponse     `json:"tenant"`
	TenantUser TenantUserResponse `json:"tenant_user"`
}

func (cfg *apiConfig) handlerTenantsCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "creating resource: tenant", "request_id", reqID)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	type parameters struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error decoding params", "error", err, "request_id", reqID)
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Tenant name is required", nil)
		return
	}

	// Generate GID for the tenant
	tenantGID := cfg.gidGen.Generate()

	// Create the tenant
	tenant, err := cfg.db.CreateTenant(r.Context(), database.CreateTenantParams{
		Gid:  sql.NullInt64{Int64: int64(tenantGID), Valid: true},
		Name: params.Name,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating tenant", "error", err, "request_id", reqID)
		respondWithError(w, http.StatusBadRequest, "Unable to create tenant", err)
		return
	}

	// Add the creating user as an active tenant user
	tenantUser, err := cfg.db.CreateTenantUser(r.Context(), database.CreateTenantUserParams{
		TenantID: tenant.ID,
		UserID:   user,
		Status:   "active",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating tenant user", "error", err, "request_id", reqID)
		respondWithError(w, http.StatusInternalServerError, "Tenant created but failed to add user", err)
		return
	}

	slog.InfoContext(r.Context(), "Tenant created successfully",
		"tenant_id", tenant.ID,
		"user_id", user,
		"request_id", reqID,
	)

	// Create Owner role for the tenant
	roleGID := cfg.gidGen.Generate()
	description := sql.NullString{String: ownerRoleDesc, Valid: true}
	role, err := cfg.db.CreateRole(r.Context(), database.CreateRoleParams{
		Gid:         sql.NullInt64{Int64: int64(roleGID), Valid: true},
		TenantID:    tenant.ID,
		Name:        ownerRoleName,
		Description: description,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant role creation failed: database error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenant.ID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to create role", err)
		return
	}
	// Assign ALL permissions to Owner role
	permissions, err := cfg.db.GetAllPermissions(r.Context())
	if err!=nil{
		slog.WarnContext(r.Context(), "tenant role creation: error fetching permissions",
				"request_id", reqID,
				"error", err,
			)
			respondWithError(w, http.StatusInternalServerError, "Unable to create get permissions for role", err)
		return
	}
	assignedPermissions := []string{}
	if len(permissions) > 0 {
		
			for _, perm := range permissions {
				err = cfg.db.AssignPermissionToRole(r.Context(), database.AssignPermissionToRoleParams{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
				if err != nil {
					slog.WarnContext(r.Context(), "tenant role creation: error assigning permission",
						"request_id", reqID,
						"role_id", role.ID,
						"permission_key", perm.Key,
						"error", err,
					)
				} else {
					assignedPermissions = append(assignedPermissions, perm.Key)
				}
			}
		
	}

	slog.InfoContext(r.Context(), "tenant role created successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenant.ID,
		"role_id", role.ID,
		"role_name", role.Name,
		"permissions_count", len(assignedPermissions),
	)

	//role created we forgot to create add the user
	err = cfg.db.AssignRoleToTenantUser(r.Context(), database.AssignRoleToTenantUserParams{
		TenantUserID: tenantUser.ID,
		RoleID:       role.ID,
	})
	if err != nil {
		slog.WarnContext(r.Context(), "tenant role creation: error assigning permission to user",
			"request_id", reqID,
			"role_id", role.ID,
			"role_key", role.Name,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to assign role to user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, CreateTenantResponse{
		Tenant: TenantResponse{
			ID:        tenant.ID,
			GID:       gid.TenantGID(uint64(tenant.Gid.Int64)).String(),
			Name:      tenant.Name,
			Status:    tenant.Status,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		},
		TenantUser: TenantUserResponse{
			ID:        tenantUser.ID,
			TenantID:  tenantUser.TenantID,
			UserID:    tenantUser.UserID,
			Status:    tenantUser.Status,
			CreatedAt: tenantUser.CreatedAt,
			UpdatedAt: tenantUser.UpdatedAt,
		},
	})
}
