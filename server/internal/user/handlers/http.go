package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/1kyryll/onetube/server/internal/auth"
	"github.com/1kyryll/onetube/server/internal/common/gen"
	s3keys "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/1kyryll/onetube/server/internal/config"
	"github.com/1kyryll/onetube/server/internal/user/types"
	"github.com/1kyryll/onetube/server/internal/util"
)

type UserHTTPHandler struct {
	cfg *config.Config
	svc types.UserService
}

func NewUserHTTPHandler(cfg *config.Config, svc types.UserService) *UserHTTPHandler {
	return &UserHTTPHandler{
		cfg: cfg,
		svc: svc,
	}
}

func (h *UserHTTPHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req types.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Username == "" || req.DisplayName == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Username, req.DisplayName, hash)
	if err != nil {
		http.Error(w, "Failed to signup user", http.StatusInternalServerError)
		return
	}

	h.setSessionCookie(w, user.ID)
	resp, err := h.toUserResponse(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to build user response", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	ok, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	h.setSessionCookie(w, user.ID)
	resp, err := h.toUserResponse(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to build user response", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHTTPHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	resp, err := h.toUserResponse(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to build user response", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHTTPHandler) toUserResponse(ctx context.Context, user *gen.User) (types.UserResponse, error) {
	resp := types.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}
	if user.AvatarKey.Valid {
		url, err := h.svc.GetAvatarUrl(ctx, user.AvatarKey.String, 15*time.Minute)
		if err != nil {
			return types.UserResponse{}, err
		}
		resp.Avatar = url
	}
	return resp, nil
}

func (h *UserHTTPHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.UploadAvatarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ContentType = strings.TrimSpace(req.ContentType)
	if !strings.HasPrefix(req.ContentType, "image/") {
		http.Error(w, "content_type must be an image/* MIME type", http.StatusBadRequest)
		return
	}

	url, key, err := h.svc.CreateAvatarUploadIntent(r.Context(), uid, req.ContentType)
	if err != nil {
		http.Error(w, "Failed to create upload intent", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, types.UploadAvatarResponse{
		Url: url,
		Key: key,
	})
}

func (h *UserHTTPHandler) CompleteAvatarUpload(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	key := s3keys.AvatarKey(uid)
	if err := h.svc.UpdateUserAvatarKey(r.Context(), uid, key); err != nil {
		http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHTTPHandler) setSessionCookie(w http.ResponseWriter, uid uuid.UUID) {
	token, _ := auth.IssueToken(h.cfg.JWTSecret, uid, 24*time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}
