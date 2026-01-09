package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoleResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var roleCursorCodec = CursorCodec[RoleCursor]{
	Validate: func(c RoleCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("invalid cursor: missing required fields")
		}
		return nil
	},
}

const (
	defaultRoleLimit = 50
	maxRoleLimit     = 100
)

// handlerTenantRolesCreate creates a new role within a tenant
func (cfg *apiConfig) handlerTenantRolesCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "tenant role creation request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	// Verify requesting user has permission to manage users (which includes role management)
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "tenant:manage_users",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant role creation failed: error checking permissions",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		slog.WarnContext(r.Context(), "tenant role creation denied: insufficient permissions",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
		)
		respondWithError(w, http.StatusForbidden, "You do not have permission to create roles", nil)
		return
	}

	type parameters struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"` // Permission keys to assign
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Role name is required", nil)
		return
	}

	// Check if role name already exists in this tenant
	_, err = cfg.db.GetRoleByTenantAndName(r.Context(), database.GetRoleByTenantAndNameParams{
		TenantID: tenantID,
		Name:     params.Name,
	})
	if err == nil {
		respondWithError(w, http.StatusConflict, "A role with this name already exists", nil)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "Unable to check existing roles", err)
		return
	}

	// Create the role
	description := sql.NullString{String: params.Description, Valid: params.Description != ""}
	role, err := cfg.db.CreateRole(r.Context(), database.CreateRoleParams{
		TenantID:    tenantID,
		Name:        params.Name,
		Description: description,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant role creation failed: database error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to create role", err)
		return
	}

	// Assign permissions if provided
	assignedPermissions := []string{}
	if len(params.Permissions) > 0 {
		permissions, err := cfg.db.GetPermissionsByKeys(r.Context(), params.Permissions)
		if err != nil {
			slog.WarnContext(r.Context(), "tenant role creation: error fetching permissions",
				"request_id", reqID,
				"error", err,
			)
		} else {
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
	}

	slog.InfoContext(r.Context(), "tenant role created successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"role_id", role.ID,
		"role_name", role.Name,
		"permissions_count", len(assignedPermissions),
	)

	var desc *string
	if role.Description.Valid {
		desc = &role.Description.String
	}

	respondWithJSON(w, http.StatusCreated, RoleResponse{
		ID:          role.ID,
		TenantID:    role.TenantID,
		Name:        role.Name,
		Description: desc,
		Permissions: assignedPermissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

// handlerTenantRolesList lists all roles in a tenant with pagination
func (cfg *apiConfig) handlerTenantRolesList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "tenant roles list request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	// Verify user is a member of the tenant
	tenantUser, err := cfg.db.GetTenantUser(r.Context(), database.GetTenantUserParams{
		TenantID: tenantID,
		UserID:   user,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusForbidden, "You are not a member of this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify membership", err)
		return
	}

	if tenantUser.Status != "active" {
		respondWithError(w, http.StatusForbidden, "Your membership is not active", nil)
		return
	}

	pageParams, err := ParsePageParams(r, defaultRoleLimit, maxRoleLimit)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	limit := pageParams.Limit
	limitPlusOne := int32(pageParams.Limit + 1)

	cursorCreatedAt, cursorID, hasCursor, err := decodeRoleCursor(pageParams.Cursor)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	rows, err := cfg.db.GetRolesByTenantIDPaginated(r.Context(), database.GetRolesByTenantIDPaginatedParams{
		TenantID: tenantID,
		Column2:  hasCursor,
		Column3:  cursorCreatedAt,
		Column4:  cursorID,
		Limit:    limitPlusOne,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant roles list failed: database query error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve roles", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = roleCursorCodec.Encode(RoleCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	// Fetch permissions for each role
	response := make([]RoleResponse, 0, len(rows))
	for _, role := range rows {
		permissions, err := cfg.db.GetPermissionsByRoleID(r.Context(), role.ID)
		if err != nil {
			slog.WarnContext(r.Context(), "tenant roles list: error fetching permissions for role",
				"request_id", reqID,
				"role_id", role.ID,
				"error", err,
			)
			permissions = []database.Permission{}
		}

		permissionKeys := make([]string, 0, len(permissions))
		for _, perm := range permissions {
			permissionKeys = append(permissionKeys, perm.Key)
		}

		var desc *string
		if role.Description.Valid {
			desc = &role.Description.String
		}

		response = append(response, RoleResponse{
			ID:          role.ID,
			TenantID:    role.TenantID,
			Name:        role.Name,
			Description: desc,
			Permissions: permissionKeys,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant roles list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"role_count", len(response),
		"has_more", hasMore,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": response,
		"page": map[string]any{
			"limit":       limit,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		},
	})
}

// handlerTenantRoleAddPermission adds a permission to a role
func (cfg *apiConfig) handlerTenantRoleAddPermission(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	roleParam := chi.URLParam(r, "roleID")

	slog.InfoContext(r.Context(), "tenant role add permission request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"role_param", roleParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	roleID, err := uuid.Parse(roleParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid role ID format", err)
		return
	}

	// Verify requesting user has permission to manage users
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "tenant:manage_users",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to manage roles", nil)
		return
	}

	// Verify role exists and belongs to this tenant
	role, err := cfg.db.GetRoleByTenantAndID(r.Context(), database.GetRoleByTenantAndIDParams{
		TenantID: tenantID,
		ID:       roleID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Role not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to find role", err)
		return
	}

	type parameters struct {
		PermissionKey string `json:"permission_key"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	if params.PermissionKey == "" {
		respondWithError(w, http.StatusBadRequest, "Permission key is required", nil)
		return
	}

	// Find the permission by key
	permission, err := cfg.db.GetPermissionByKey(r.Context(), params.PermissionKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Permission not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to find permission", err)
		return
	}

	// Assign permission to role
	err = cfg.db.AssignPermissionToRole(r.Context(), database.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permission.ID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant role add permission failed: database error",
			"request_id", reqID,
			"role_id", roleID,
			"permission_id", permission.ID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to add permission to role", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant role permission added successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"role_id", roleID,
		"role_name", role.Name,
		"permission_key", permission.Key,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":        "Permission added to role successfully",
		"role_id":        roleID,
		"role_name":      role.Name,
		"permission_key": permission.Key,
	})
}

// handlerTenantRoleRemovePermission removes a permission from a role
func (cfg *apiConfig) handlerTenantRoleRemovePermission(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	roleParam := chi.URLParam(r, "roleID")
	permissionParam := chi.URLParam(r, "permissionKey")

	slog.InfoContext(r.Context(), "tenant role remove permission request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"role_param", roleParam,
		"permission_param", permissionParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	roleID, err := uuid.Parse(roleParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid role ID format", err)
		return
	}

	// Verify requesting user has permission to manage users
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "tenant:manage_users",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to manage roles", nil)
		return
	}

	// Find the permission by key
	permission, err := cfg.db.GetPermissionByKey(r.Context(), permissionParam)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Permission not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to find permission", err)
		return
	}

	// Remove permission from role
	err = cfg.db.RemovePermissionFromRole(r.Context(), database.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permission.ID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant role remove permission failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to remove permission from role", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant role permission removed successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"role_id", roleID,
		"permission_key", permissionParam,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":        "Permission removed from role successfully",
		"role_id":        roleID,
		"permission_key": permissionParam,
	})
}

// handlerPermissionsList lists all available permissions
func (cfg *apiConfig) handlerPermissionsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	slog.InfoContext(r.Context(), "permissions list request received",
		"request_id", reqID,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	permissions, err := cfg.db.GetAllPermissions(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "permissions list failed: database query error",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve permissions", err)
		return
	}

	response := make([]PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		var desc *string
		if perm.Description.Valid {
			desc = &perm.Description.String
		}

		response = append(response, PermissionResponse{
			ID:          perm.ID,
			Key:         perm.Key,
			Description: desc,
			CreatedAt:   perm.CreatedAt,
			UpdatedAt:   perm.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "permissions list successful",
		"request_id", reqID,
		"user_id", user,
		"permission_count", len(response),
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": response,
	})
}

func decodeRoleCursor(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := roleCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}
