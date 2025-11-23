package store

import (
	"RealTime/internal/domain/user" // <-- 1. FIXED: Import the correct domain package
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("store: resource not found")

// UserStore implements the user.Storer interface for PostgreSQL.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new UserStore.
// This implements the user.Storer interface
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

// Create inserts a new user into the database.
func (s *UserStore) Create(ctx context.Context, u *user.User) error {
	query := `INSERT INTO users (id, username, hashed_password, created_at)
              VALUES ($1, $2, $3, $4)`
	//idAsUUID := uuid.UUID(u.ID)
	_, err := s.db.ExecContext(ctx, query, u.ID, u.Username, u.HashedPassword, u.CreatedAt)
	if err != nil {
		// Here you could check for specific pq errors, like duplicate key
		return fmt.Errorf("failed to execute user creation query: %w", err)
	}
	return nil
}

// GetByUsername retrieves a user by their username.
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `SELECT id, username, hashed_password, created_at FROM users WHERE username = $1`

	u := &user.User{}

	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID,
		&u.Username,
		&u.HashedPassword,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return u, nil
}
