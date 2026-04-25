package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/1kyryll/onetube/server/internal/auth"
	"github.com/1kyryll/onetube/server/internal/config"
	"github.com/1kyryll/onetube/server/internal/util"
	"github.com/1kyryll/onetube/server/internal/video/services"
	"github.com/1kyryll/onetube/server/internal/video/types"
)

type VideoHTTPHandler struct {
	cfg *config.Config
	svc types.VideoService
}

func NewVideoHTTPHandler(cfg *config.Config, svc types.VideoService) *VideoHTTPHandler {
	return &VideoHTTPHandler{
		cfg: cfg,
		svc: svc,
	}
}

func (h *VideoHTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req types.CreateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Filename = strings.TrimSpace(req.Filename)
	if req.Title == "" || req.Filename == "" {
		http.Error(w, "title and filename are required", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.CreateUploadIntent(r.Context(), uid, req)
	if err != nil {
		http.Error(w, "Failed to create upload intent", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, http.StatusCreated, resp)
}

func (h *VideoHTTPHandler) Complete(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}

	var req types.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkUploaded(r.Context(), uid, vid, req.Extension); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to mark uploaded", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *VideoHTTPHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.GetForWatch(r.Context(), vid)
	if err != nil {
		if errors.Is(err, services.ErrNotReady) {
			util.WriteJSON(w, http.StatusOK, map[string]string{"status": "processing"})
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	util.WriteJSON(w, http.StatusOK, resp)
}

func (h *VideoHTTPHandler) Feed(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 24, 100)
	cursor := parseCursor(r.URL.Query().Get("cursor"))

	page, err := h.svc.ListFeed(r.Context(), limit, cursor)
	if err != nil {
		http.Error(w, "Failed to list feed", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, page)
}

func (h *VideoHTTPHandler) MyVideos(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"), 24, 100)
	cursor := parseCursor(r.URL.Query().Get("cursor"))

	page, err := h.svc.ListByUploader(r.Context(), uid, limit, cursor)
	if err != nil {
		http.Error(w, "Failed to list videos", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, page)
}

func (h *VideoHTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}

	if err := h.svc.SoftDelete(r.Context(), uid, vid); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoHTTPHandler) MasterPlaylist(w http.ResponseWriter, r *http.Request) {
	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}
	body, err := h.svc.GetMasterPlaylist(r.Context(), vid)
	if err != nil {
		if errors.Is(err, services.ErrNotReady) {
			http.Error(w, "Not ready", http.StatusConflict)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	writePlaylist(w, body)
}

func (h *VideoHTTPHandler) RenditionPlaylist(w http.ResponseWriter, r *http.Request) {
	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}
	quality := r.PathValue("quality")
	if quality == "" || strings.ContainsAny(quality, "/\\") {
		http.Error(w, "Invalid quality", http.StatusBadRequest)
		return
	}
	body, err := h.svc.GetRenditionPlaylist(r.Context(), vid, quality)
	if err != nil {
		if errors.Is(err, services.ErrNotReady) {
			http.Error(w, "Not ready", http.StatusConflict)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	writePlaylist(w, body)
}

func writePlaylist(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(body)
}

func (h *VideoHTTPHandler) IncrementViewCount(w http.ResponseWriter, r *http.Request) {
	vid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid video id", http.StatusBadRequest)
		return
	}
	if err := h.svc.IncrementViewCount(r.Context(), vid); err != nil {
		http.Error(w, "Failed to record view", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseLimit(raw string, defaultVal, max int) int32 {
	if raw == "" {
		return int32(defaultVal)
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return int32(defaultVal)
	}
	if n > max {
		return int32(max)
	}
	return int32(n)
}

func parseCursor(raw string) *types.FeedCursor {
	if raw == "" {
		return nil
	}

	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return nil
	}

	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil
	}
	return &types.FeedCursor{PublishedAt: ts, ID: id}
}
