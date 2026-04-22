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

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
}
