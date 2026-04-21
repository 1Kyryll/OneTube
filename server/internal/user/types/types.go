package types

import (
	"context"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*gen.User, error)
	GetUserByEmail(ctx context.Context, email string) (*gen.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*gen.User, error)
}
