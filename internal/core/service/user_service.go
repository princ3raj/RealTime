package service

import (
	domain "RealTime/internal/domain/user"
	"context"
	"errors"
	"fmt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username is already taken")
)

// Storer defines the contract for data storage.
// Our service will depend on this, not a concrete *sql.DB.
type Storer interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

// Service orchestrates the business logic.
// It uses the Storer interface to interact with the database.
type Service struct {
	store Storer
}

// NewService is the constructor for our user service.
func NewUserService(store Storer) *Service {
	return &Service{
		store: store,
	}
}

// Register handles the logic for new user registration.
func (s *Service) Register(ctx context.Context, username, password string) (*domain.User, error) {

	existingUser, _ := s.store.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}

	user, err := domain.NewUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	if err := s.store.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// Login handles user authentication.
func (s *Service) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := user.ComparePassword(password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
