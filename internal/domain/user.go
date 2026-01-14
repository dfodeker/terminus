// internal/domain/user.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID            uuid.UUID
	Email         string
	EmailVerified bool
	PasswordHash  string
	FirstName     string
	LastName      string
	AvatarURL     string
	Phone         string
	Status        string
	Locale        string
	Timezone      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	DeletedAt     *time.Time
}

func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return ""
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

func (u *User) IsActive() bool {
	return u.Status == string(UserStatusActive) && u.DeletedAt == nil
}
