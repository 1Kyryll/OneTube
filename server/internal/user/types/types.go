package types

import (
	"context"

	"github.com/1kyryll/onetube/server/internal/common/gen"
)

type UserService interface {
	Signup(ctx context.Context, email, username, displayName, password string) (*gen.User, error)
	Login(ctx context.Context, email, password string) (*gen.User, error)
}
