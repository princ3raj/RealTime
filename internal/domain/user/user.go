package user

import (
	"RealTime/internal/types"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username is already taken")
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

// Storer defines the contract for data storage.
// Our service will depend on this, not a concrete *sql.DB.
type Storer interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	// You might add: GetByID(ctx context.Context, id ulid.ULID) (*User, error)
}

// Service orchestrates the business logic.
// It uses the Storer interface to interact with the database.
type Service struct {
	store Storer
	// You could add a logger here if needed
}

// NewService is the constructor for our user service.
func NewService(store Storer) *Service {
	return &Service{
		store: store,
	}
}

// Register handles the logic for new user registration.
func (s *Service) Register(ctx context.Context, username, password string) (*User, error) {
	// 1. Check if user already exists
	// We deliberately ignore the error here, we only care if the user *exists*
	existingUser, _ := s.store.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}

	// 2. Create the new user entity
	user, err := NewUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	// 3. Persist to storage
	if err := s.store.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// Login handles user authentication.
func (s *Service) Login(ctx context.Context, username, password string) (*User, error) {
	// 1. Find the user by username
	user, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		// This could be a db error OR a "not found" error.
		// In either case, from a security standpoint, it's invalid credentials.
		return nil, ErrInvalidCredentials
	}

	// 2. Compare the password
	if err := user.ComparePassword(password); err != nil {
		// Password does not match
		return nil, ErrInvalidCredentials
	}

	// 3. Success
	return user, nil
}
