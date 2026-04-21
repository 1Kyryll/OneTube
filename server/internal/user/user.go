package services

import (
	"context"

	"github.com/1kyryll/onetube/server/internal/common/gen"
)

type UserServiceImpl struct {
	queries *gen.Queries
}

func NewUserService(queries *gen.Queries) *UserServiceImpl {
	return &UserServiceImpl{
		queries: queries,
	}
}

func (s *UserServiceImpl) Signup(ctx context.Context, email, username, displayName, password string) (*gen.User, error) {
	return &gen.User{}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, email, password string) (*gen.User, error) {
	return &gen.User{}, nil
}
