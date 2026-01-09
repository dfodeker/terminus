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

type TenantMemberResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantMemberCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var tenantMemberCursorCodec = CursorCodec[TenantMemberCursor]{
	Validate: func(c TenantMemberCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("invalid cursor: missing required fields")
		}
		return nil
	},
}

const (
	defaultMemberLimit = 50
	maxMemberLimit     = 100
)

// handlerTenantMembersInvite invites a user to a tenant
func (cfg *apiConfig) handlerTenantMembersInvite(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "tenant member invite request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "tenant member invite failed: no authenticated user",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant member invite failed: invalid tenant ID",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	// Verify requesting user has permission to invite
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "tenant:invite_users",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant member invite failed: error checking permissions",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		slog.WarnContext(r.Context(), "tenant member invite denied: insufficient permissions",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
		)
		respondWithError(w, http.StatusForbidden, "You do not have permission to invite users to this tenant", nil)
		return
	}

	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant member invite failed: invalid request body",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	// Find user by email
	invitedUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.WarnContext(r.Context(), "tenant member invite failed: user not found",
				"request_id", reqID,
				"user_id", user,
				"tenant_id", tenantID,
				"invited_email", params.Email,
			)
			respondWithError(w, http.StatusNotFound, "User with this email not found", nil)
			return
		}
		slog.ErrorContext(r.Context(), "tenant member invite failed: error finding user",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to find user", err)
		return
	}

	// Check if user is already a member
	existingMember, err := cfg.db.GetTenantUser(r.Context(), database.GetTenantUserParams{
		TenantID: tenantID,
		UserID:   invitedUser.ID,
	})
	if err == nil {
		if existingMember.Status == "active" {
			respondWithError(w, http.StatusConflict, "User is already an active member of this tenant", nil)
			return
		}
		if existingMember.Status == "invited" {
			respondWithError(w, http.StatusConflict, "User has already been invited to this tenant", nil)
			return
		}
		// If status is 'removed', we can re-invite
	} else if !errors.Is(err, sql.ErrNoRows) {
		slog.ErrorContext(r.Context(), "tenant member invite failed: error checking existing membership",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to check existing membership", err)
		return
	}

	// Create tenant user with 'invited' status
	tenantUser, err := cfg.db.CreateTenantUser(r.Context(), database.CreateTenantUserParams{
		TenantID: tenantID,
		UserID:   invitedUser.ID,
		Status:   "invited",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant member invite failed: error creating tenant user",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"invited_user_id", invitedUser.ID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to create invitation", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant member invited successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"invited_user_id", invitedUser.ID,
		"tenant_user_id", tenantUser.ID,
	)

	respondWithJSON(w, http.StatusCreated, TenantMemberResponse{
		ID:        tenantUser.ID,
		TenantID:  tenantUser.TenantID,
		UserID:    tenantUser.UserID,
		Email:     invitedUser.Email,
		Status:    tenantUser.Status,
		Roles:     []string{},
		CreatedAt: tenantUser.CreatedAt,
		UpdatedAt: tenantUser.UpdatedAt,
	})
}

// handlerTenantMembersList lists all members of a tenant with pagination
func (cfg *apiConfig) handlerTenantMembersList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "tenant members list request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "tenant members list failed: no authenticated user",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant members list failed: invalid tenant ID",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
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
			slog.WarnContext(r.Context(), "tenant members list denied: user not a member",
				"request_id", reqID,
				"user_id", user,
				"tenant_id", tenantID,
			)
			respondWithError(w, http.StatusForbidden, "You are not a member of this tenant", nil)
			return
		}
		slog.ErrorContext(r.Context(), "tenant members list failed: error checking membership",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify membership", err)
		return
	}

	if tenantUser.Status != "active" {
		respondWithError(w, http.StatusForbidden, "Your membership is not active", nil)
		return
	}

	pageParams, err := ParsePageParams(r, defaultMemberLimit, maxMemberLimit)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	limit := pageParams.Limit
	limitPlusOne := int32(pageParams.Limit + 1)

	cursorCreatedAt, cursorID, hasCursor, err := decodeTenantMemberCursor(pageParams.Cursor)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	slog.DebugContext(r.Context(), "tenant members list: fetching members from database",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"limit", limit,
		"has_cursor", hasCursor,
	)

	rows, err := cfg.db.GetTenantUsersWithDetailsPaginated(r.Context(), database.GetTenantUsersWithDetailsPaginatedParams{
		TenantID: tenantID,
		Column2:  hasCursor,
		Column3:  cursorCreatedAt,
		Column4:  cursorID,
		Limit:    limitPlusOne,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant members list failed: database query error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve members", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = tenantMemberCursorCodec.Encode(TenantMemberCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	// Fetch roles for each member
	response := make([]TenantMemberResponse, 0, len(rows))
	for _, member := range rows {
		roles, err := cfg.db.GetRolesByTenantUserID(r.Context(), member.ID)
		if err != nil {
			slog.ErrorContext(r.Context(), "tenant members list: error fetching roles for member",
				"request_id", reqID,
				"tenant_user_id", member.ID,
				"error", err,
			)
			roles = []database.Role{}
		}

		roleNames := make([]string, 0, len(roles))
		for _, role := range roles {
			roleNames = append(roleNames, role.Name)
		}

		response = append(response, TenantMemberResponse{
			ID:        member.ID,
			TenantID:  member.TenantID,
			UserID:    member.UserID,
			Email:     member.Email,
			Status:    member.Status,
			Roles:     roleNames,
			CreatedAt: member.CreatedAt,
			UpdatedAt: member.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant members list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"member_count", len(response),
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

// handlerTenantMemberAssignRole assigns a role to a tenant member
func (cfg *apiConfig) handlerTenantMemberAssignRole(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	memberParam := chi.URLParam(r, "memberID")

	slog.InfoContext(r.Context(), "tenant member role assignment request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"member_param", memberParam,
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

	memberID, err := uuid.Parse(memberParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid member ID format", err)
		return
	}

	// Verify requesting user has permission to manage users
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "tenant:manage_users",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant member role assignment failed: error checking permissions",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		slog.WarnContext(r.Context(), "tenant member role assignment denied: insufficient permissions",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
		)
		respondWithError(w, http.StatusForbidden, "You do not have permission to manage users", nil)
		return
	}

	// Verify the member exists and belongs to this tenant
	tenantUser, err := cfg.db.GetTenantUserByID(r.Context(), memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Member not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to find member", err)
		return
	}

	if tenantUser.TenantID != tenantID {
		respondWithError(w, http.StatusNotFound, "Member not found in this tenant", nil)
		return
	}

	type parameters struct {
		RoleID string `json:"role_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	roleID, err := uuid.Parse(params.RoleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid role ID format", err)
		return
	}

	// Verify the role exists and belongs to this tenant
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

	// Assign the role
	err = cfg.db.AssignRoleToTenantUser(r.Context(), database.AssignRoleToTenantUserParams{
		TenantUserID: memberID,
		RoleID:       roleID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant member role assignment failed: database error",
			"request_id", reqID,
			"tenant_user_id", memberID,
			"role_id", roleID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to assign role", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant member role assigned successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"tenant_user_id", memberID,
		"role_id", roleID,
		"role_name", role.Name,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":   "Role assigned successfully",
		"member_id": memberID,
		"role_id":   roleID,
		"role_name": role.Name,
	})
}

// handlerTenantMemberRemoveRole removes a role from a tenant member
func (cfg *apiConfig) handlerTenantMemberRemoveRole(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	memberParam := chi.URLParam(r, "memberID")
	roleParam := chi.URLParam(r, "roleID")

	slog.InfoContext(r.Context(), "tenant member role removal request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"member_param", memberParam,
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

	memberID, err := uuid.Parse(memberParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid member ID format", err)
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
		respondWithError(w, http.StatusForbidden, "You do not have permission to manage users", nil)
		return
	}

	// Remove the role
	err = cfg.db.RemoveRoleFromTenantUser(r.Context(), database.RemoveRoleFromTenantUserParams{
		TenantUserID: memberID,
		RoleID:       roleID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant member role removal failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to remove role", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant member role removed successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"tenant_user_id", memberID,
		"role_id", roleID,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":   "Role removed successfully",
		"member_id": memberID,
		"role_id":   roleID,
	})
}

func decodeTenantMemberCursor(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := tenantMemberCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}
