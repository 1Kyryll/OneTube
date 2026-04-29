package services

import (
	"context"
	"time"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	s3client "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserServiceImpl struct {
	queries *gen.Queries
	s3      *s3client.Client
}

func NewUserService(queries *gen.Queries, s3 *s3client.Client) *UserServiceImpl {
	return &UserServiceImpl{
		queries: queries,
		s3:      s3,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, email, username, displayName, passwordHash string, avatarKey string) (*gen.User, error) {
	u, err := s.queries.CreateUser(ctx, gen.CreateUserParams{
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		AvatarKey:    pgtype.Text{String: avatarKey, Valid: true},
	})

	return &u, err
}

func (s *UserServiceImpl) CreateAvatarUploadIntent(ctx context.Context, userID uuid.UUID, contentType string) (string, string, error) {
	key := s3client.AvatarKey(userID)
	url, err := s.s3.PresignPut(ctx, key, contentType, 15*time.Minute)

	return url, key, err
}

func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*gen.User, error) {
	u, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*gen.User, error) {
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
