package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Seb0sti0n/astrophage/backend/internal/analysis"
	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	"github.com/Seb0sti0n/astrophage/backend/internal/llm"
	"github.com/Seb0sti0n/astrophage/backend/internal/seed"
)

const testSecret = "test-secret"

// testEnv is an API server backed by a freshly seeded PostgreSQL database. The tests use their
// own database (astrophage_test), never the development one, because the seed truncates it.
type testEnv struct {
	t      *testing.T
	store  *db.Store
	router http.Handler
	token  string
}

func testDatabaseURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/astrophage_test?sslmode=disable"
}

func newEnv(t *testing.T, stepDelay time.Duration) *testEnv {
	t.Helper()
	return newEnvLLM(t, stepDelay, config.LLM{})
}

// newEnvLLM is newEnv with an LLM configuration (an empty one disables the LLM).
func newEnvLLM(t *testing.T, stepDelay time.Duration, llmCfg config.LLM) *testEnv {
	t.Helper()
	ctx := context.Background()
	url := testDatabaseURL()

	// Create the test database if it does not exist yet (connect to the maintenance DB first).
	admin, err := pgx.Connect(ctx, strings.Replace(url, "/astrophage_test", "/postgres", 1))
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	var exists bool
	_ = admin.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'astrophage_test')`).Scan(&exists)
	if !exists {
		if _, err := admin.Exec(ctx, `CREATE DATABASE astrophage_test`); err != nil {
			t.Fatal(err)
		}
	}
	admin.Close(ctx)

	if err := seed.Migrate(url); err != nil {
		t.Fatal(err)
	}
	pool, err := db.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := seed.Run(ctx, pool, "../../../data"); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{JWTSecret: testSecret, CORSOrigin: "http://localhost:5173", StepDelay: stepDelay, Engine: config.DefaultEngine(), LLM: llmCfg}
	store := db.NewStore(pool)
	env := &testEnv{t: t, store: store, router: NewServer(store, analysis.NewRunner(store, cfg, llm.New(cfg.LLM)), cfg).Router()}

	code, body := env.do("POST", "/api/auth/login", `{"email":"demo@energy.io","password":"demo123"}`, "")
	if code != 200 {
		t.Fatalf("login failed: %d %s", code, body)
	}
	env.token = decode[struct{ Token string }](t, body).Token
	return env
}

func (e *testEnv) do(method, path, body, token string) (int, []byte) {
	e.t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

// get/send call an authenticated endpoint.
func (e *testEnv) get(path string) (int, []byte)        { return e.do("GET", path, "", e.token) }
func (e *testEnv) send(m, path, b string) (int, []byte) { return e.do(m, path, b, e.token) }

func decode[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("bad JSON %q: %v", b, err)
	}
	return v
}

func (e *testEnv) mustGet(path string, want int) []byte {
	e.t.Helper()
	code, body := e.get(path)
	if code != want {
		e.t.Fatalf("GET %s = %d, want %d: %s", path, code, want, body)
	}
	return body
}

type run struct {
	ID          int
	Status      string
	CurrentStep *string `json:"current_step"`
	Steps       []struct{ Name, Status string }
	Summary     struct {
		Anomalies      int
		HighPriority   int `json:"high_priority"`
		MetersAnalyzed int `json:"meters_analyzed"`
		ReadingsCount  int `json:"readings_count"`
		EventsCount    int `json:"events_count"`
	}
}

// analyze starts an analysis and waits for it to finish.
func (e *testEnv) analyze() run {
	e.t.Helper()
	code, body := e.send("POST", "/api/ai/analyze", "")
	if code != http.StatusAccepted {
		e.t.Fatalf("analyze = %d: %s", code, body)
	}
	return e.waitFor(decode[struct{ ID int }](e.t, body).ID)
}

func (e *testEnv) waitFor(id int) run {
	e.t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		r := decode[run](e.t, e.mustGet("/api/ai/analysis/"+itoa(id), 200))
		if r.Status == "COMPLETED" || r.Status == "FAILED" {
			return r
		}
		time.Sleep(20 * time.Millisecond)
	}
	e.t.Fatalf("analysis %d did not finish in time", id)
	return run{}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

type meterRow struct {
	MeterID      string  `json:"meter_id"`
	Status       string  `json:"status"`
	Consumption  float64 `json:"consumption_kwh"`
	Baseline     float64 `json:"baseline_kwh"`
	VariationPct float64 `json:"variation_pct"`
	Anomaly      *struct{ Type, Severity string }
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	DailyKWh     []float64 `json:"daily_kwh"`
	DailyFrom    string    `json:"daily_from"`
}

// meterPageResp is the paginated meters list.
type meterPageResp struct {
	Items    []meterRow     `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Counts   map[string]int `json:"counts"`
}

// page fetches one page of meters with the given query string (without a leading "?").
func (e *testEnv) page(query string) meterPageResp {
	e.t.Helper()
	path := "/api/meters"
	if query != "" {
		path += "?" + query
	}
	return decode[meterPageResp](e.t, e.mustGet(path, 200))
}

// meterList fetches every meter matching the query, using the largest page.
func (e *testEnv) meterList(query string) []meterRow {
	e.t.Helper()
	q := "page_size=100"
	if query != "" {
		q = query + "&" + q
	}
	return e.page(q).Items
}

func meterIDs(ms []meterRow) []string {
	out := []string{}
	for _, m := range ms {
		out = append(out, m.MeterID)
	}
	return out
}

func eq(a, b []string) bool { return strings.Join(a, ",") == strings.Join(b, ",") }

func TestAuth(t *testing.T) {
	e := newEnv(t, 0)
	signed := func(secret string, exp time.Time) string {
		s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "x", ExpiresAt: jwt.NewNumericDate(exp)}).SignedString([]byte(secret))
		return s
	}
	tests := []struct {
		name  string
		token string
		want  int
	}{
		{"no token", "", 401},
		{"garbage token", "abc.def.ghi", 401},
		{"signed with another secret", signed("other", time.Now().Add(time.Hour)), 401},
		{"expired token", signed(testSecret, time.Now().Add(-time.Hour)), 401},
		{"valid token", signed(testSecret, time.Now().Add(time.Hour)), 200},
		{"token from login", e.token, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code, _ := e.do("GET", "/api/meters", "", tt.token); code != tt.want {
				t.Errorf("got %d, want %d", code, tt.want)
			}
		})
	}
	for name, body := range map[string]string{
		"wrong password": `{"email":"demo@energy.io","password":"nope"}`,
		"unknown user":   `{"email":"nobody@energy.io","password":"demo123"}`,
		"invalid body":   `not json`,
	} {
		code, _ := e.do("POST", "/api/auth/login", body, "")
		want := 401
		if name == "invalid body" {
			want = 400
		}
		if code != want {
			t.Errorf("%s: got %d, want %d", name, code, want)
		}
	}
	// health and login are public
	if code, _ := e.do("GET", "/api/health", "", ""); code != 200 {
		t.Errorf("health = %d", code)
	}
}

func TestBeforeAnyAnalysis(t *testing.T) {
	e := newEnv(t, 0)

	ms := e.meterList("")
	if len(ms) != 12 {
		t.Fatalf("got %d meters, want 12", len(ms))
	}
	for _, m := range ms {
		// Nobody has analysed anything yet: the API must not claim the meters are fine.
		if m.Status != "UNEVALUATED" || m.Anomaly != nil {
			t.Errorf("%s: status %s anomaly %v before analysis, want UNEVALUATED and none", m.MeterID, m.Status, m.Anomaly)
		}
	}
	for status, want := range map[string]int{"unevaluated": 12, "ok": 0, "alert": 0, "critical": 0} {
		if got := e.meterList("status=" + status); len(got) != want {
			t.Errorf("status=%s before analysis returned %d meters, want %d", status, len(got), want)
		}
	}
	if d := decode[meterRow](t, e.mustGet("/api/meters/M-109", 200)); d.Status != "UNEVALUATED" {
		t.Errorf("M-109 detail status = %s before analysis", d.Status)
	}
	byVar := e.meterList("sort=variation")
	if byVar[0].MeterID != "M-109" || byVar[0].VariationPct < 105 || byVar[0].VariationPct > 115 {
		t.Errorf("first by variation = %+v, want M-109 ≈ +110%%", byVar[0])
	}
	if byVar[1].MeterID != "M-104" {
		t.Errorf("second by variation = %s, want M-104", byVar[1].MeterID)
	}
	// Variation sorts by size of the change, up or down: M-110 (-0.1%) is the smallest, not M-105 (-1.4%).
	smallest := e.meterList("sort=variation&order=asc")
	if smallest[0].MeterID != "M-110" {
		t.Errorf("smallest variation = %s, want M-110 (sorted by absolute value)", smallest[0].MeterID)
	}
	byID := map[string]meterRow{}
	for _, m := range ms {
		byID[m.MeterID] = m
	}
	if m := byID["M-109"]; m.Name != "Tablero principal B" || m.Location != "Planta Sur" {
		t.Errorf("M-109 name/location = %q / %q, want the design's", m.Name, m.Location)
	}
	if m := byID["M-101"]; m.Name != "Compresores A" || m.Location != "Planta Norte" {
		t.Errorf("M-101 name/location = %q / %q", m.Name, m.Location)
	}
	if got := byID["M-101"].DailyFrom; got != "2026-09-01" {
		t.Errorf("daily_from = %q, want 2026-09-01", got)
	}
	for _, m := range ms {
		if len(m.DailyKWh) != 14 {
			t.Fatalf("%s has %d daily points, want 14", m.MeterID, len(m.DailyKWh))
		}
	}
	if d := byID["M-109"].DailyKWh; d[13] < 2200 || d[13] > 2215 || d[0] > 1100 {
		t.Errorf("M-109 daily series = %v", d)
	}
	asc := e.meterList("sort=consumption&order=asc")
	if asc[0].MeterID != "M-107" || asc[11].MeterID != "M-109" {
		t.Errorf("consumption asc = %v", meterIDs(asc))
	}
	if got := meterIDs(e.meterList("search=11")); !eq(got, []string{"M-110", "M-111", "M-112"}) {
		t.Errorf("search 11 = %v", got)
	}
	if got := e.meterList("search=zzz"); len(got) != 0 {
		t.Errorf("search zzz = %v", got)
	}
	for _, bad := range []string{"status=nope", "sort=nope", "order=up"} {
		e.mustGet("/api/meters?"+bad, 400)
	}

	e.mustGet("/api/ai/analysis/latest", 404)
	e.mustGet("/api/ai/analysis/999", 404)
	e.mustGet("/api/ai/analysis/abc", 400)
	if got := decode[[]any](t, e.mustGet("/api/anomalies", 200)); len(got) != 0 {
		t.Errorf("anomalies before analysis = %v", got)
	}

	d := decode[struct {
		Meters      int                     `json:"meters_count"`
		Total       float64                 `json:"total_consumption_kwh"`
		Count       int                     `json:"anomalies_count"`
		AvgConf     *float64                `json:"avg_confidence"`
		LastRun     *struct{}               `json:"last_analysis"`
		Daily       []struct{ Date string } `json:"daily_consumption"`
		TopPriority []any                   `json:"top_priorities"`
	}](t, e.mustGet("/api/dashboard/summary", 200))
	if d.Meters != 12 || d.Count != 0 || d.AvgConf != nil || d.LastRun != nil || len(d.Daily) != 14 || len(d.TopPriority) != 0 {
		t.Errorf("summary before analysis = %+v", d)
	}
	if d.Total < 155000 || d.Total > 155500 {
		t.Errorf("total consumption = %.1f, want ≈ 155.251", d.Total)
	}
}

// The full demo cycle: analyze → poll → anomalies, with priority order and derived meter status.
func TestAnalysisCycle(t *testing.T) {
	e := newEnv(t, 0)
	r := e.analyze()

	if r.Status != "COMPLETED" || r.CurrentStep != nil {
		t.Fatalf("run = %+v", r)
	}
	if len(r.Steps) != 7 || r.Steps[0].Name != "Lecturas" || r.Steps[6].Name != "Recomendación" {
		t.Fatalf("steps = %+v", r.Steps)
	}
	for _, s := range r.Steps {
		if s.Status != "DONE" {
			t.Errorf("step %s is %s", s.Name, s.Status)
		}
	}
	if r.Summary.ReadingsCount != 4032 || r.Summary.EventsCount != 4 {
		t.Errorf("counts = %d readings, %d events, want 4032 and 4", r.Summary.ReadingsCount, r.Summary.EventsCount)
	}
	if r.Summary.Anomalies != 4 || r.Summary.HighPriority != 2 || r.Summary.MetersAnalyzed != 12 {
		t.Errorf("summary = %+v", r.Summary)
	}
	latest := decode[run](t, e.mustGet("/api/ai/analysis/latest", 200))
	if latest.ID != r.ID {
		t.Errorf("latest = %d, want %d", latest.ID, r.ID)
	}

	type anomaly struct {
		ID            int
		MeterID       string `json:"meter_id"`
		Type          string
		Severity      string
		Confidence    float64
		PriorityScore float64 `json:"priority_score"`
		Status        string
		Source        string `json:"explanation_source"`
		Reason        string
		Explanation   string
		Action        string `json:"recommended_action"`
		Evidence      map[string]any
	}
	list := decode[[]anomaly](t, e.mustGet("/api/anomalies", 200))
	want := []struct{ meter, typ, sev string }{
		{"M-109", "REAL_ANOMALY", "HIGH"}, {"M-112", "DATA_QUALITY", "HIGH"},
		{"M-104", "EXPLAINABLE_ANOMALY", "MEDIUM"}, {"M-106", "FALSE_POSITIVE", "LOW"},
	}
	if len(list) != 4 {
		t.Fatalf("got %d anomalies", len(list))
	}
	for i, w := range want {
		a := list[i]
		if a.MeterID != w.meter || a.Type != w.typ || a.Severity != w.sev || a.Status != "OPEN" || a.Source != "TEMPLATE" {
			t.Errorf("anomaly %d = %+v, want %+v", i, a, w)
		}
		if a.Reason == "" || a.Explanation == "" || a.Action == "" || a.Confidence <= 0 {
			t.Errorf("anomaly %s is missing text or confidence", a.MeterID)
		}
		if i > 0 && a.PriorityScore > list[i-1].PriorityScore {
			t.Errorf("not sorted by priority at %d", i)
		}
	}

	// Filters.
	for path, want := range map[string][]string{
		"/api/anomalies?severity=high":                   {"M-109", "M-112"},
		"/api/anomalies?type=false_positive":             {"M-106"},
		"/api/anomalies?type=DATA_QUALITY&severity=HIGH": {"M-112"},
		"/api/anomalies?severity=low&type=REAL_ANOMALY":  {},
		"/api/anomalies?status=resolved":                 {},
	} {
		got := []string{}
		for _, a := range decode[[]anomaly](t, e.mustGet(path, 200)) {
			got = append(got, a.MeterID)
		}
		if !eq(got, want) {
			t.Errorf("%s = %v, want %v", path, got, want)
		}
	}
	e.mustGet("/api/anomalies?type=nope", 400)

	// Detail includes the evidence.
	d := decode[anomaly](t, e.mustGet("/api/anomalies/"+itoa(list[0].ID), 200))
	for _, k := range []string{"window", "metrics", "changed_variables", "events", "rules_fired", "priority_breakdown", "confidence_breakdown"} {
		if _, ok := d.Evidence[k]; !ok {
			t.Errorf("evidence is missing %q", k)
		}
	}
	e.mustGet("/api/anomalies/99999", 404)
	e.mustGet("/api/anomalies/abc", 400)

	// Meter statuses are derived from the anomalies; filters and severity sort use them.
	if left := e.meterList("status=unevaluated"); len(left) != 0 {
		t.Errorf("%d meters still unevaluated after an analysis", len(left))
	}
	crit := e.meterList("status=critical")
	alert := e.meterList("status=alert")
	ok := e.meterList("status=ok")
	if !eq(meterIDs(crit), []string{"M-109"}) || !eq(meterIDs(alert), []string{"M-104", "M-112"}) || len(ok) != 9 {
		t.Errorf("critical=%v alert=%v ok=%d", meterIDs(crit), meterIDs(alert), len(ok))
	}
	for _, m := range ok {
		if m.MeterID == "M-106" && (m.Anomaly == nil || m.Anomaly.Type != "FALSE_POSITIVE") {
			t.Errorf("M-106 should be OK but still show its explained anomaly: %+v", m)
		}
	}
	sorted := e.meterList("sort=severity")
	if !eq(meterIDs(sorted)[:4], []string{"M-109", "M-112", "M-104", "M-106"}) {
		t.Errorf("severity order = %v", meterIDs(sorted))
	}
	detail := decode[meterRow](t, e.mustGet("/api/meters/M-109", 200))
	if detail.Status != "CRITICAL" || detail.Anomaly == nil || detail.Anomaly.Type != "REAL_ANOMALY" {
		t.Errorf("M-109 detail = %+v", detail)
	}
	e.mustGet("/api/meters/M-999", 404)

	// Dashboard.
	s := decode[struct {
		Count      int            `json:"anomalies_count"`
		High       int            `json:"high_priority_count"`
		AvgConf    *float64       `json:"avg_confidence"`
		Last       *run           `json:"last_analysis"`
		HasResults bool           `json:"has_results"`
		HighMeters []string       `json:"high_priority_meters"`
		ByType     map[string]int `json:"anomalies_by_type"`
		Top        []struct {
			MeterID string `json:"meter_id"`
		} `json:"top_priorities"`
	}](t, e.mustGet("/api/dashboard/summary", 200))
	if !s.HasResults {
		t.Errorf("has_results must be true after a completed analysis")
	}
	if !eq(s.HighMeters, []string{"M-109", "M-112"}) || s.ByType["REAL_ANOMALY"] != 1 || s.ByType["DATA_QUALITY"] != 1 ||
		s.ByType["EXPLAINABLE_ANOMALY"] != 1 || s.ByType["FALSE_POSITIVE"] != 1 {
		t.Errorf("high meters = %v, by type = %v", s.HighMeters, s.ByType)
	}
	if s.Count != 4 || s.High != 2 || s.AvgConf == nil || *s.AvgConf < 0.85 || s.Last == nil || s.Last.Status != "COMPLETED" || len(s.Top) != 4 || s.Top[0].MeterID != "M-109" {
		t.Errorf("summary = %+v", s)
	}

	// Acknowledge / resolve.
	id := "/api/anomalies/" + itoa(list[0].ID)
	for _, status := range []string{"ACKNOWLEDGED", "RESOLVED"} {
		code, body := e.send("PATCH", id, `{"status":"`+status+`"}`)
		if code != 200 || decode[anomaly](t, body).Status != status {
			t.Errorf("patch %s = %d %s", status, code, body)
		}
	}
	if got := decode[[]anomaly](t, e.mustGet("/api/anomalies?status=resolved", 200)); len(got) != 1 {
		t.Errorf("resolved anomalies = %d", len(got))
	}
	if code, _ := e.send("PATCH", id, `{"status":"DONE"}`); code != 400 {
		t.Errorf("invalid status = %d", code)
	}
	if code, _ := e.send("PATCH", id, `nope`); code != 400 {
		t.Errorf("invalid body = %d", code)
	}
	if code, _ := e.send("PATCH", "/api/anomalies/99999", `{"status":"OPEN"}`); code != 404 {
		t.Errorf("patch missing = %d", code)
	}

	// Running again replaces the visible anomalies with the new run's: still 4, all OPEN again.
	e.analyze()
	again := decode[[]anomaly](t, e.mustGet("/api/anomalies", 200))
	if len(again) != 4 || again[0].Status != "OPEN" || again[0].ID == list[0].ID {
		t.Errorf("after a second run: %d anomalies, first %+v", len(again), again[0])
	}
}

func TestOnlyOneAnalysisAtATime(t *testing.T) {
	e := newEnv(t, 150*time.Millisecond)

	code, body := e.send("POST", "/api/ai/analyze", "")
	if code != http.StatusAccepted {
		t.Fatalf("first analyze = %d", code)
	}
	first := decode[struct{ ID int }](t, body).ID

	code, body = e.send("POST", "/api/ai/analyze", "")
	if code != http.StatusConflict || decode[struct{ ID int }](t, body).ID != first {
		t.Errorf("second analyze = %d %s, want 409 with id %d", code, body, first)
	}

	// While it runs the progress is visible: a current step and some steps still pending.
	time.Sleep(200 * time.Millisecond)
	mid := decode[run](t, e.mustGet("/api/ai/analysis/"+itoa(first), 200))
	if mid.Status != "RUNNING" || mid.CurrentStep == nil {
		t.Errorf("mid-run = %+v", mid)
	}
	pending := 0
	for _, s := range mid.Steps {
		if s.Status == "PENDING" {
			pending++
		}
	}
	if pending == 0 {
		t.Errorf("expected pending steps mid-run, got %+v", mid.Steps)
	}

	if r := e.waitFor(first); r.Status != "COMPLETED" {
		t.Fatalf("run = %+v", r)
	}
	// Once finished, a new one can start.
	if code, _ := e.send("POST", "/api/ai/analyze", ""); code != http.StatusAccepted {
		t.Errorf("analyze after completion = %d", code)
	}
	time.Sleep(1200 * time.Millisecond) // let it finish before the next test reseeds
}

func TestStaleRunsAreFailedOnStartup(t *testing.T) {
	e := newEnv(t, 0)
	ctx := context.Background()
	id, err := e.store.CreateRun(ctx, []byte("[]"))
	if err != nil {
		t.Fatal(err)
	}
	// A run stuck in progress blocks new analyses...
	if code, _ := e.send("POST", "/api/ai/analyze", ""); code != http.StatusConflict {
		t.Fatalf("analyze with a stuck run = %d, want 409", code)
	}
	// ...until startup cleanup marks it failed.
	if err := e.store.FailStaleRuns(ctx); err != nil {
		t.Fatal(err)
	}
	if r := decode[run](t, e.mustGet("/api/ai/analysis/"+itoa(id), 200)); r.Status != "FAILED" {
		t.Errorf("stale run = %s", r.Status)
	}
	// A failed run produced no results, so the meters are still unevaluated.
	if d := decode[meterRow](t, e.mustGet("/api/meters/M-109", 200)); d.Status != "UNEVALUATED" {
		t.Errorf("status after a failed run = %s, want UNEVALUATED", d.Status)
	}
	if r := e.analyze(); r.Status != "COMPLETED" {
		t.Errorf("new run = %s", r.Status)
	}
}

func TestReadingsAndEvents(t *testing.T) {
	e := newEnv(t, 0)
	type point struct {
		Timestamp   time.Time
		Consumption float64 `json:"consumption_kwh"`
		Baseline    float64 `json:"baseline_kwh"`
		Low         float64 `json:"band_low_kwh"`
		High        float64 `json:"band_high_kwh"`
		BaselinePF  float64 `json:"baseline_power_factor"`
		PowerFactor float64 `json:"power_factor"`
	}
	type series struct {
		MeterID     string `json:"meter_id"`
		Granularity string
		Points      []point
	}

	hourly := decode[series](t, e.mustGet("/api/meters/M-109/readings", 200))
	if hourly.Granularity != "hour" || len(hourly.Points) != 336 {
		t.Fatalf("hourly = %s, %d points", hourly.Granularity, len(hourly.Points))
	}
	if !hourly.Points[0].Timestamp.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("first point at %v", hourly.Points[0].Timestamp)
	}
	for _, p := range hourly.Points {
		if !(p.Low <= p.Baseline && p.Baseline <= p.High) {
			t.Fatalf("band does not contain the baseline at %v: %+v", p.Timestamp, p)
		}
	}

	// The anomaly window starts at 12 Sep 14:00, well outside the baseline band.
	win := decode[series](t, e.mustGet("/api/meters/M-109/readings?from=2026-09-12T13:00:00Z&to=2026-09-12T14:00:00Z", 200))
	if len(win.Points) != 2 {
		t.Fatalf("range = %d points", len(win.Points))
	}
	before, after := win.Points[0], win.Points[1]
	if before.Consumption > before.High*1.2 || after.Consumption < after.High*1.5 {
		t.Errorf("expected a jump above the band at 14:00: before=%+v after=%+v", before, after)
	}
	if after.PowerFactor > after.BaselinePF-0.15 {
		t.Errorf("expected the power factor to drop: %+v", after)
	}
	if got := decode[series](t, e.mustGet("/api/meters/M-109/readings?from=2026-09-14&to=2026-09-14", 200)); len(got.Points) != 1 {
		t.Errorf("date-only range = %d points, want 1", len(got.Points))
	}

	daily := decode[series](t, e.mustGet("/api/meters/M-109/readings?granularity=day", 200))
	if len(daily.Points) != 14 {
		t.Fatalf("daily = %d points", len(daily.Points))
	}
	last := daily.Points[13]
	if last.Consumption < 2200 || last.Consumption > 2215 || last.Baseline < 1040 || last.Baseline > 1055 || last.High < last.Baseline {
		t.Errorf("last day = %+v", last)
	}
	if last.PowerFactor > 0.8 || daily.Points[0].PowerFactor < 0.9 {
		t.Errorf("daily power factor should be averaged: first %.3f last %.3f", daily.Points[0].PowerFactor, last.PowerFactor)
	}

	for _, bad := range []string{"granularity=week", "from=yesterday", "to=13/09"} {
		e.mustGet("/api/meters/M-109/readings?"+bad, 400)
	}
	e.mustGet("/api/meters/M-999/readings", 404)

	events := decode[[]struct{ Type, Description string }](t, e.mustGet("/api/meters/M-109/events", 200))
	if len(events) != 1 || events[0].Type != "UNKNOWN" {
		t.Errorf("M-109 events = %+v", events)
	}
	if got := decode[[]any](t, e.mustGet("/api/meters/M-101/events", 200)); len(got) != 0 {
		t.Errorf("M-101 events = %v", got)
	}
	e.mustGet("/api/meters/M-999/events", 404)
}

func TestMetersPagination(t *testing.T) {
	e := newEnv(t, 0)
	// A stable order across pages: without an explicit sort it is by meter_id.
	all := e.meterList("")
	if len(all) != 12 {
		t.Fatalf("got %d meters", len(all))
	}

	t.Run("defaults to 10 per page and says how many there are in total", func(t *testing.T) {
		p := e.page("")
		if p.Page != 1 || p.PageSize != 10 || p.Total != 12 || len(p.Items) != 10 {
			t.Errorf("page=%d size=%d total=%d items=%d, want 1, 10, 12, 10", p.Page, p.PageSize, p.Total, len(p.Items))
		}
	})

	t.Run("walks through every meter exactly once", func(t *testing.T) {
		var seen []string
		for page := 1; page <= 3; page++ {
			p := e.page("page_size=5&page=" + itoa(page))
			if p.Total != 12 || p.Page != page || p.PageSize != 5 {
				t.Errorf("page %d: %+v", page, p)
			}
			seen = append(seen, meterIDs(p.Items)...)
		}
		if len(seen) != 12 || !eq(seen, meterIDs(all)) {
			t.Errorf("pages joined = %v, want %v", seen, meterIDs(all))
		}
	})

	t.Run("the last page is partial and a page past the end is empty, not an error", func(t *testing.T) {
		if got := len(e.page("page_size=5&page=3").Items); got != 2 {
			t.Errorf("last page has %d items, want 2", got)
		}
		p := e.page("page_size=5&page=9")
		if len(p.Items) != 0 || p.Total != 12 {
			t.Errorf("page 9: %d items, total %d", len(p.Items), p.Total)
		}
	})

	t.Run("sorting applies to the whole list before it is cut into pages", func(t *testing.T) {
		// Highest consumption overall is M-109 (2208 kWh), then M-104: they lead page 1, not "the top of each page".
		first := e.page("sort=consumption&order=desc&page_size=2&page=1")
		second := e.page("sort=consumption&order=desc&page_size=2&page=2")
		if !eq(meterIDs(first.Items), []string{"M-109", "M-104"}) {
			t.Errorf("page 1 by consumption = %v", meterIDs(first.Items))
		}
		if second.Items[0].Consumption > first.Items[1].Consumption {
			t.Errorf("page 2 starts higher than page 1 ends: %v", meterIDs(second.Items))
		}
	})

	t.Run("total follows the filters, counts do not", func(t *testing.T) {
		p := e.page("search=11&page_size=2")
		if p.Total != 3 || len(p.Items) != 2 {
			t.Errorf("search=11: total %d, %d items, want 3 and 2", p.Total, len(p.Items))
		}
		if p.Counts["ALL"] != 12 || p.Counts["UNEVALUATED"] != 12 {
			t.Errorf("counts must cover all meters regardless of the search: %v", p.Counts)
		}
	})

	t.Run("counts by status after an analysis", func(t *testing.T) {
		e.analyze()
		p := e.page("status=alert&page_size=1")
		if p.Total != 2 || len(p.Items) != 1 {
			t.Errorf("status=alert: total %d, %d items, want 2 and 1", p.Total, len(p.Items))
		}
		want := map[string]int{"ALL": 12, "OK": 9, "ALERT": 2, "CRITICAL": 1, "UNEVALUATED": 0}
		for k, v := range want {
			if p.Counts[k] != v {
				t.Errorf("counts[%s] = %d, want %d (%v)", k, p.Counts[k], v, p.Counts)
			}
		}
	})

	t.Run("rejects invalid paging parameters", func(t *testing.T) {
		for _, bad := range []string{"page=0", "page=-1", "page=x", "page_size=0", "page_size=101", "page_size=x", "page_size=1.5"} {
			e.mustGet("/api/meters?"+bad, 400)
		}
		e.mustGet("/api/meters?page_size=100", 200) // the largest allowed
	})
}
