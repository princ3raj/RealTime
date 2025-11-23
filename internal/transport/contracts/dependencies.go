package contracts

import (
	"RealTime/internal/config"
	"RealTime/internal/domain/user"
	"context"
)

type UserServiceProvider interface {
	Register(ctx context.Context, username, password string) (*user.User, error)
	Login(ctx context.Context, username, password string) (*user.User, error)
}

type Publisher interface {
	Publish(topic string, message []byte) error
}

type AppDependencies struct {
	UserService UserServiceProvider
	Publisher   Publisher
	Config      *config.Config
}
