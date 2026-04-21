package http

import "net/http"

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", s.HandleSignup)
	mux.HandleFunc("/login", s.Login)
	mux.HandleFunc("/logout", s.Logout)
	mux.Handle("/me", s.requireAuth(http.HandlerFunc(s.GetCurrentUser)))
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
