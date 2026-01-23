// internal/domain/oauth_installation.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type OAuthInstallation struct {
	ID              uuid.UUID
	ShopID          uuid.UUID
	AppID           uuid.UUID
	AccessTokenHash string
	Scopes          []Scope
	Status          InstallationStatus
	InstalledAt     time.Time
	UninstalledAt   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InstallationStatus string

const (
	InstallationStatusActive      InstallationStatus = "active"
	InstallationStatusUninstalled InstallationStatus = "uninstalled"
	InstallationStatusSuspended   InstallationStatus = "suspended"
)
