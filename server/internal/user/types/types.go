package types

import (
	"context"
	"time"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*gen.User, error)
	GetUserByEmail(ctx context.Context, email string) (*gen.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*gen.User, error)
	CreateAvatarUploadIntent(ctx context.Context, userID uuid.UUID, contentType string) (string, string, error)
	UpdateUserAvatarKey(ctx context.Context, userID uuid.UUID, avatarKey string) error
	GetAvatarUrl(ctx context.Context, avatarKey string, ttl time.Duration) (string, error)
}

type SignupRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UploadAvatarRequest struct {
	ContentType string `json:"content_type"`
}

type UploadAvatarResponse struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Avatar      string    `json:"avatar"`
}
