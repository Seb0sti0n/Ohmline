package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) analyze(w http.ResponseWriter, r *http.Request) {
	id, inProgress, err := s.runner.Start(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	if inProgress {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "an analysis is already running", "id": id})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]int{"id": id})
}

func (s *Server) getAnalysis(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	run, err := s.store.Run(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if run == nil {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) latestAnalysis(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.LatestRun(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	if run == nil {
		writeError(w, http.StatusNotFound, "no analysis has been run yet")
		return
	}
	writeJSON(w, http.StatusOK, run)
}
