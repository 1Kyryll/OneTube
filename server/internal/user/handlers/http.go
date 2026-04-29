package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/1kyryll/onetube/server/internal/auth"
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
	avatarKey := ""

	if req.Email == "" || req.Username == "" || req.DisplayName == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Username, req.DisplayName, hash, avatarKey)
	if err != nil {
		http.Error(w, "Failed to signup user", http.StatusInternalServerError)
		return
	}

	h.setSessionCookie(w, user.ID)
	util.WriteJSON(w, http.StatusCreated, types.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
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
	util.WriteJSON(w, http.StatusOK, types.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
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

	util.WriteJSON(w, http.StatusOK, types.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
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

	if req.Avatar == "" {
		http.Error(w, "No avatar assigned", http.StatusBadRequest)
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(req.Avatar)
	if err != nil {
		http.Error(w, "Invalid base64 avatar", http.StatusBadRequest)
		return
	}

	contentType := http.DetectContentType(imgBytes)
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "Avatar is not an image", http.StatusBadRequest)
		return
	}

	url, key, err := h.svc.CreateAvatarUploadIntent(r.Context(), uid, contentType)
	if err != nil {
		http.Error(w, "Failed to upload avatar", http.StatusInternalServerError)
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

	var req types.CompleteAvatarUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "No avatar key assigned", http.StatusBadRequest)
		return
	}

	h.svc.UpdateUserAvatarKey(r.Context(), uid, req.Key)

	w.WriteHeader(http.StatusOK)
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
