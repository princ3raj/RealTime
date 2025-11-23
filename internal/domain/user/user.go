package user

import (
	"RealTime/internal/types"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system.
type User struct {
	ID             types.SQLULID `json:"id"`
	Username       string        `json:"username"`
	HashedPassword []byte        `json:"-"` // Correctly hidden
	CreatedAt      time.Time     `json:"created_at"`
}

// NewUser is a factory for creating a new User.
// This enforces business rules (e.g., password hashing).
func NewUser(username, password string) (*User, error) {
	// You might add validation here:
	if username == "" || password == "" {
		return nil, errors.New("username and password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.Reader, 0)
	newID := ulid.MustNew(ulid.Timestamp(t), entropy)

	return &User{
		ID:             types.SQLULID{ULID: newID},
		Username:       username,
		HashedPassword: hashedPassword,
		CreatedAt:      t,
	}, nil
}

// ComparePassword checks if the provided password matches the user's hash.
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
}
