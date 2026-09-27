package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Seb0sti0n/Ohmline/backend/internal/db"
)

var (
	anomalyTypes    = set("REAL_ANOMALY", "EXPLAINABLE_ANOMALY", "FALSE_POSITIVE", "DATA_QUALITY")
	severities      = set("HIGH", "MEDIUM", "LOW")
	anomalyStatuses = set("OPEN", "ACKNOWLEDGED", "RESOLVED")
)

func set(vals ...string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vals {
		m[v] = true
	}
	return m
}

// queryEnum reads an optional upper-cased enum filter; ok is false (and a 400 written) if invalid.
func queryEnum(w http.ResponseWriter, r *http.Request, name string, allowed map[string]bool) (string, bool) {
	v := strings.ToUpper(r.URL.Query().Get(name))
	if v != "" && !allowed[v] {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return "", false
	}
	return v, true
}

func (s *Server) listAnomalies(w http.ResponseWriter, r *http.Request) {
	var f db.AnomalyFilter
	var ok bool
	if f.Type, ok = queryEnum(w, r, "type", anomalyTypes); !ok {
		return
	}
	if f.Severity, ok = queryEnum(w, r, "severity", severities); !ok {
		return
	}
	if f.Status, ok = queryEnum(w, r, "status", anomalyStatuses); !ok {
		return
	}
	list, err := s.store.Anomalies(r.Context(), f)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func (s *Server) getAnomaly(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.Anomaly(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if a == nil {
		writeError(w, http.StatusNotFound, "anomaly not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) patchAnomaly(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&body); err != nil || !anomalyStatuses[body.Status] {
		writeError(w, http.StatusBadRequest, "status must be OPEN, ACKNOWLEDGED or RESOLVED")
		return
	}
	found, err := s.store.UpdateAnomalyStatus(r.Context(), id, body.Status)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "anomaly not found")
		return
	}
	a, err := s.store.Anomaly(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}
