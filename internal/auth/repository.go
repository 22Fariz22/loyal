package auth

import (
	"context"

	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/pkg/logger"
)

type UserRepository interface {
	CreateUser(ctx context.Context, l logger.Interface, user *entity.User) error
	GetUser(ctx context.Context, l logger.Interface, username, password string) (*entity.User, error)
}
