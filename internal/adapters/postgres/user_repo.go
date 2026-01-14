// internal/adapters/postgres/user_repo.go
package postgres

import (
	"context"
	"errors"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type UserRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		PasswordHash:  textToPgtype(user.PasswordHash),
		FirstName:     textToPgtype(user.FirstName),
		LastName:      textToPgtype(user.LastName),
		AvatarUrl:     textToPgtype(user.AvatarURL),
		Phone:         textToPgtype(user.Phone),
		Status:        user.Status,
		Locale:        user.Locale,
		Timezone:      user.Timezone,
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return err
	}

	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return dbUserToDomain(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return dbUserToDomain(row), nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        user.ID,
		FirstName: textToPgtype(user.FirstName),
		LastName:  textToPgtype(user.LastName),
		AvatarUrl: textToPgtype(user.AvatarURL),
		Phone:     textToPgtype(user.Phone),
		Locale:    textToPgtype(user.Locale),
		Timezone:  textToPgtype(user.Timezone),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteUser(ctx, id)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: textToPgtype(passwordHash),
	})
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateLastLogin(ctx, id)
}

func (r *UserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.VerifyUserEmail(ctx, id)
	return err
}

func dbUserToDomain(row db.User) *domain.User {
	user := &domain.User{
		ID:            row.ID,
		Email:         row.Email,
		EmailVerified: row.EmailVerified,
		Status:        row.Status,
		Locale:        row.Locale,
		Timezone:      row.Timezone,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	if row.PasswordHash.Valid {
		user.PasswordHash = row.PasswordHash.String
	}
	if row.FirstName.Valid {
		user.FirstName = row.FirstName.String
	}
	if row.LastName.Valid {
		user.LastName = row.LastName.String
	}
	if row.AvatarUrl.Valid {
		user.AvatarURL = row.AvatarUrl.String
	}
	if row.Phone.Valid {
		user.Phone = row.Phone.String
	}
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		user.LastLoginAt = &t
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		user.DeletedAt = &t
	}

	return user
}

