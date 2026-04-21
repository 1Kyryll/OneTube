package types

import (
	"context"

	"github.com/1kyryll/onetube/server/internal/common/gen"
)

type UserService interface {
	Signup(ctx context.Context, email, username, displayName, password string) (*gen.User, error)
}
