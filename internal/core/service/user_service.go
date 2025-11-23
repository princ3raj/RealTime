package service

import (
	domain "RealTime/internal/domain/user" // Alias this clearly
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log" // Or use your internal logger if available
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username is already taken")
)

// Publisher defines the output port for events.
type Publisher interface {
	Publish(topic string, message []byte) error
}

// Storer defines the contract for data storage.
type Storer interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

// UserService orchestrates the business logic.
type UserService struct {
	store     Storer
	publisher Publisher
}

func NewUserService(store Storer, pub Publisher) *UserService {
	return &UserService{
		store:     store,
		publisher: pub,
	}
}

// Register handles creation AND event publishing (Atomic Business Logic)
func (s *UserService) Register(ctx context.Context, username, password string) (*domain.User, error) {

	// 1. Check if user exists
	existingUser, _ := s.store.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}

	// 2. Create Domain Entity (Factory)
	user, err := domain.NewUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	// 3. Persist to Database
	if err := s.store.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 4. Side Effect: Publish Event
	// We do this AFTER the DB save is successful.
	if err := s.publishUserRegistered(user); err != nil {
		log.Printf("ERROR: Failed to publish user_registered event for %s: %v", user.ID.String(), err)
	}

	return user, nil
}

// Helper to keep the main logic clean
func (s *UserService) publishUserRegistered(u *domain.User) error {
	eventData, err := json.Marshal(map[string]string{
		"type":    "USER_REGISTERED",
		"user_id": u.ID.String(),
		"email":   u.Username,
	})
	if err != nil {
		return err
	}
	return s.publisher.Publish("user_events", eventData)
}

// Login handles user authentication.
func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := user.ComparePassword(password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
