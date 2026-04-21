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

func (s *UserServiceImpl) CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*gen.User, error) {
	u, err := s.queries.CreateUser(ctx, gen.CreateUserParams{
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
	})

	return &u, err
}

func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*gen.User, error) {
	u, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
