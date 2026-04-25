package main

import (
	"net/http"

	"github.com/1kyryll/onetube/server/internal/auth"
	"github.com/1kyryll/onetube/server/internal/config"
	userhandlers "github.com/1kyryll/onetube/server/internal/user/handlers"
	videohandlers "github.com/1kyryll/onetube/server/internal/video/handlers"
)

func newRouter(cfg *config.Config, userH *userhandlers.UserHTTPHandler, videoH *videohandlers.VideoHTTPHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/signup", userH.Signup)
	mux.HandleFunc("POST /api/auth/login", userH.Login)
	mux.HandleFunc("POST /api/auth/logout", userH.Logout)
	mux.Handle("GET /api/auth/me", auth.RequireAuth(cfg, http.HandlerFunc(userH.GetCurrentUser)))

	mux.Handle("POST /api/videos", auth.RequireAuth(cfg, http.HandlerFunc(videoH.Create)))
	mux.Handle("POST /api/videos/{id}/complete", auth.RequireAuth(cfg, http.HandlerFunc(videoH.Complete)))
	mux.Handle("DELETE /api/videos/{id}", auth.RequireAuth(cfg, http.HandlerFunc(videoH.Delete)))
	mux.Handle("GET /api/videos/mine", auth.RequireAuth(cfg, http.HandlerFunc(videoH.MyVideos)))

	mux.HandleFunc("GET /api/videos/feed", videoH.Feed)
	mux.HandleFunc("GET /api/videos/{id}", videoH.GetStatus)
	mux.HandleFunc("POST /api/videos/{id}/view", videoH.IncrementViewCount)
	mux.HandleFunc("GET /api/videos/{id}/hls/master.m3u8", videoH.MasterPlaylist)
	mux.HandleFunc("GET /api/videos/{id}/hls/{quality}/playlist.m3u8", videoH.RenditionPlaylist)

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
