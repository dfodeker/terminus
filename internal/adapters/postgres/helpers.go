// internal/adapters/postgres/helpers.go
package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// textToPgtype converts a string to pgtype.Text
func textToPgtype(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// uuidToPgtype converts a uuid.UUID to pgtype.UUID
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// uuidPtrToPgtype converts a *uuid.UUID to pgtype.UUID
func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// timePtrToPgtype converts a *time.Time to pgtype.Timestamptz
func timePtrToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// int64ToPgtype converts an int64 to pgtype.Int8
func int64ToPgtype(i int64) pgtype.Int8 {
	if i == 0 {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: i, Valid: true}
}

// isPgUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isPgUniqueViolation(err error) bool {
	// Check for pgx error code 23505 (unique_violation)
	var pgErr interface{ Code() string }
	if errors.As(err, &pgErr) {
		return pgErr.Code() == "23505"
	}
	return false
}
