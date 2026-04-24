package main

import (
	"net/http"

	"github.com/1kyryll/onetube/server/internal/auth"
	"github.com/1kyryll/onetube/server/internal/config"
	videohandlers "github.com/1kyryll/onetube/server/internal/video/handlers"
)

func newRouter(cfg *config.Config, h *videohandlers.VideoHTTPHandler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/videos", auth.RequireAuth(cfg, http.HandlerFunc(h.Create)))
	mux.Handle("POST /api/videos/{id}/complete", auth.RequireAuth(cfg, http.HandlerFunc(h.Complete)))
	mux.Handle("DELETE /api/videos/{id}", auth.RequireAuth(cfg, http.HandlerFunc(h.Delete)))
	mux.Handle("GET /api/videos/mine", auth.RequireAuth(cfg, http.HandlerFunc(h.MyVideos)))

	mux.HandleFunc("GET /api/videos/feed", h.Feed)
	mux.HandleFunc("GET /api/videos/{id}", h.GetStatus)
	mux.HandleFunc("POST /api/videos/{id}/view", h.IncrementViewCount)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
