// Package http exposes the REST API under /api.
package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Seb0sti0n/astrophage/backend/internal/analysis"
	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
)

type Server struct {
	store  *db.Store
	runner *analysis.Runner
	cfg    config.Config
}

func NewServer(store *db.Store, runner *analysis.Runner, cfg config.Config) *Server {
	return &Server{store: store, runner: runner, cfg: cfg}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{s.cfg.CORSOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.health)
		r.Post("/auth/login", s.login)

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)
			r.Get("/meters", s.listMeters)
			r.Get("/meters/{meterId}", s.getMeter)
			r.Get("/meters/{meterId}/readings", s.getReadings)
			r.Get("/meters/{meterId}/events", s.getMeterEvents)

			r.Get("/anomalies", s.listAnomalies)
			r.Get("/anomalies/{id}", s.getAnomaly)
			r.Patch("/anomalies/{id}", s.patchAnomaly)

			r.Post("/ai/analyze", s.analyze)
			r.Get("/ai/analysis/latest", s.latestAnalysis)
			r.Get("/ai/analysis/{id}", s.getAnalysis)

			r.Get("/dashboard/summary", s.dashboardSummary)
		})
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status, code := "ok", http.StatusOK
	if err := s.store.Ping(r.Context()); err != nil {
		status, code = "down", http.StatusServiceUnavailable
	}
	writeJSON(w, code, map[string]string{"status": status, "db": status})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// serverError logs the real error and answers with a generic 500.
func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
