// Package web expõe a API e a interface do Incidente Zero (iz ui).
package web

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/progress"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

type Server struct {
	ScenariosDir string
	Store        *progress.Store
	Static       fs.FS // conteúdo de web/dist
	Version      string

	labMu sync.Mutex // um start/stop por vez
}

// Handler monta as rotas da API e do frontend.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.status)
	mux.HandleFunc("GET /api/scenarios", s.list)
	mux.HandleFunc("GET /api/scenarios/{id}", s.detail)
	mux.HandleFunc("POST /api/scenarios/{id}/start", s.start)
	mux.HandleFunc("POST /api/scenarios/{id}/stop", s.stop)
	mux.HandleFunc("POST /api/scenarios/{id}/check", s.check)
	mux.HandleFunc("GET /api/scenarios/{id}/steps", s.steps)
	mux.HandleFunc("GET /api/scenarios/{id}/metrics", s.metrics)
	mux.HandleFunc("POST /api/scenarios/{id}/hints/{n}", s.hint)
	mux.HandleFunc("GET /api/scenarios/{id}/solution", s.solution)
	mux.HandleFunc("GET /api/scenarios/{id}/terminal", s.terminal)
	mux.Handle("/", s.spa())
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; font-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string, extra ...map[string]any) {
	body := map[string]any{"error": msg}
	for _, e := range extra {
		for k, v := range e {
			body[k] = v
		}
	}
	writeJSON(w, code, body)
}

func (s *Server) find(w http.ResponseWriter, r *http.Request) (*scenario.Scenario, bool) {
	sc, err := scenario.Find(s.ScenariosDir, r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return nil, false
	}
	return sc, true
}

func (s *Server) findRunning(w http.ResponseWriter, r *http.Request) (*scenario.Scenario, bool) {
	sc, ok := s.find(w, r)
	if !ok {
		return nil, false
	}
	if !lab.IsRunning(sc) {
		writeErr(w, http.StatusConflict, "o lab deste cenário não está rodando")
		return nil, false
	}
	return sc, true
}

// ---------- DTOs ----------

type summary struct {
	ID          string   `json:"id"`
	Number      string   `json:"number"`
	Title       string   `json:"title"`
	Level       string   `json:"level"`
	Track       string   `json:"track"`
	Minutes     int      `json:"minutes"`
	Severity    string   `json:"severity"`
	AlertTitle  string   `json:"alertTitle"`
	Tags        []string `json:"tags"`
	HintCount   int      `json:"hintCount"`
	Running     bool     `json:"running"`
	InProgress  bool     `json:"inProgress"` // lab rodando com tentativa aberta (ainda não resolvida)
	Resolved    bool     `json:"resolved"`
	BestSeconds int      `json:"bestSeconds"`
	BestPoints  int      `json:"bestPoints"`
}

func number(id string) string {
	n, _, _ := strings.Cut(id, "-")
	return n
}

func (s *Server) summarize(sc *scenario.Scenario, running map[string]bool) summary {
	rec := s.Store.Get(sc.ID)
	hc := 0
	if h, err := sc.Hints(); err == nil {
		hc = len(h)
	}
	return summary{
		ID: sc.ID, Number: number(sc.ID), Title: sc.Title, Level: sc.Level, Track: sc.Track,
		Minutes: sc.EstimatedMinutes, Severity: sc.Alert.Severity, AlertTitle: sc.Alert.Title,
		Tags: sc.Tags, HintCount: hc, Running: running[sc.ID],
		InProgress: running[sc.ID] && rec.StartedAt != nil,
		Resolved:   rec.Resolved, BestSeconds: rec.BestSeconds, BestPoints: rec.BestPoints,
	}
}

func runningSet() map[string]bool {
	ids, _ := lab.Running()
	m := map[string]bool{}
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// ---------- handlers ----------

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	dockerOK := exec.Command("docker", "info", "--format", "{{.ServerVersion}}").Run() == nil
	runtime := ""
	if dockerOK {
		runtime = lab.Runtime()
	}
	ids, _ := lab.Running()
	all, _ := scenario.LoadAll(s.ScenariosDir)
	writeJSON(w, http.StatusOK, map[string]any{
		"version": s.Version, "docker": dockerOK, "runtime": runtime,
		"running": ids, "scenarios": len(all),
	})
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	all, err := scenario.LoadAll(s.ScenariosDir)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	running := runningSet()
	items := make([]summary, 0, len(all))
	byID := map[string]summary{}
	next := ""
	resolved, totalSecs, hints, points, maxPoints := 0, 0, 0, 0, 0
	for _, sc := range all {
		sm := s.summarize(sc, running)
		items = append(items, sm)
		byID[sc.ID] = sm
		base := sc.Scoring.Base
		if base <= 0 {
			base = 100
		}
		rec := s.Store.Get(sc.ID)
		if rec.Resolved {
			resolved++
			totalSecs += rec.LastSeconds
			hints += rec.LastHints
			points += rec.BestPoints
			maxPoints += base
		} else if next == "" {
			next = sc.ID
		}
	}

	type hist struct {
		progress.HistoryEntry
		Number   string `json:"number"`
		Title    string `json:"title"`
		Severity string `json:"severity"`
	}
	var history []hist
	for _, h := range s.Store.History() {
		sm, ok := byID[h.ID]
		if !ok {
			continue
		}
		history = append(history, hist{HistoryEntry: h, Number: sm.Number, Title: sm.Title, Severity: sm.Severity})
	}

	mttr := 0
	if resolved > 0 {
		mttr = totalSecs / resolved
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"scenarios": items,
		"next":      next,
		"history":   history,
		"stats": map[string]int{
			"resolved": resolved, "total": len(all), "mttrSeconds": mttr,
			"hintsUsed": hints, "points": points, "maxPoints": maxPoints,
		},
	})
}

func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	briefing, _ := sc.ReadFile(sc.Alert.Briefing)
	concepts, _ := sc.ReadFile(sc.Content.Concepts)
	hints, _ := sc.Hints()
	rec := s.Store.Get(sc.ID)

	type step struct {
		Title     string `json:"title"`
		Text      string `json:"text"`
		Optional  bool   `json:"optional"`
		HasVerify bool   `json:"hasVerify"`
	}
	steps := make([]step, 0, len(sc.Guided))
	for _, g := range sc.Guided {
		steps = append(steps, step{g.Title, g.Text, g.Optional, strings.TrimSpace(g.Verify) != ""})
	}
	type hintDTO struct {
		N    int    `json:"n"`
		Cost int    `json:"cost"`
		Text string `json:"text,omitempty"`
	}
	hs := make([]hintDTO, 0, len(hints))
	for i, h := range hints {
		d := hintDTO{N: i + 1, Cost: h.Cost}
		if i < rec.HintsUsed {
			d.Text = strings.TrimSpace(h.Text)
		}
		hs = append(hs, d)
	}

	var started *int64
	if rec.StartedAt != nil {
		ms := rec.StartedAt.UnixMilli()
		started = &ms
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"scenario":       s.summarize(sc, runningSet()),
		"hostname":       sc.Hostname,
		"user":           sc.User,
		"objectives":     sc.Objectives,
		"briefing":       briefing,
		"concepts":       concepts,
		"guided":         steps,
		"hints":          hs,
		"startedAt":      started,
		"solutionViewed": rec.SolutionViewed,
	})
}

func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	s.labMu.Lock()
	defer s.labMu.Unlock()

	// Um incidente por vez
	for id := range runningSet() {
		if id == sc.ID {
			continue
		}
		if other, err := scenario.Find(s.ScenariosDir, id); err == nil {
			_ = lab.Stop(other)
		}
	}

	t0 := time.Now()
	if err := lab.Start(sc); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, step := range []struct{ name, script string }{
		{"setup", sc.Scripts.Setup}, {"inject", sc.Scripts.Inject},
	} {
		out, code, err := lab.RunScript(sc, step.script)
		if err != nil || code != 0 {
			msg := step.name + " falhou"
			if err != nil {
				msg += ": " + err.Error()
			}
			writeErr(w, http.StatusInternalServerError, msg, map[string]any{"output": out, "code": code})
			return
		}
	}
	if err := s.Store.Start(sc.ID); err != nil {
		log.Printf("progresso: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ms": time.Since(t0).Milliseconds()})
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	s.labMu.Lock()
	defer s.labMu.Unlock()
	if err := lab.Stop(sc); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type checkItem struct {
	OK   bool   `json:"ok"`
	Text string `json:"text"`
}

func parseCheck(out string) []checkItem {
	var items []checkItem
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "✅"):
			items = append(items, checkItem{true, strings.TrimSpace(strings.TrimPrefix(line, "✅"))})
		case strings.HasPrefix(line, "❌"):
			items = append(items, checkItem{false, strings.TrimSpace(strings.TrimPrefix(line, "❌"))})
		}
	}
	return items
}

func (s *Server) check(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.findRunning(w, r)
	if !ok {
		return
	}
	out, code, err := lab.RunScript(sc, sc.Scripts.Check)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]any{"passed": code == 0, "items": parseCheck(out), "output": out, "at": time.Now().UnixMilli()}
	if code == 0 {
		before := s.Store.Get(sc.ID)
		if before.StartedAt != nil || !before.Resolved {
			rec, err := s.Store.Resolve(sc.ID, sc.Scoring.Base, sc.Scoring.TimeBonusMinutes)
			if err != nil {
				log.Printf("progresso: %v", err)
			}
			resp["result"] = map[string]any{
				"seconds": rec.LastSeconds, "hints": rec.LastHints,
				"points": pointsOf(before, sc), "solutionViewed": before.SolutionViewed,
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// pointsOf recalcula a pontuação da tentativa (a mesma regra de progress.Resolve).
func pointsOf(before progress.Record, sc *scenario.Scenario) int {
	base := sc.Scoring.Base
	if base <= 0 {
		base = 100
	}
	if before.SolutionViewed {
		return 0
	}
	p := base - before.HintCost
	if before.StartedAt != nil && sc.Scoring.TimeBonusMinutes > 0 &&
		time.Since(*before.StartedAt) <= time.Duration(sc.Scoring.TimeBonusMinutes)*time.Minute {
		p += 10
	}
	if p < 0 {
		p = 0
	}
	return p
}

func (s *Server) steps(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.findRunning(w, r)
	if !ok {
		return
	}
	res, err := lab.StepStatus(sc)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"steps": res})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.findRunning(w, r)
	if !ok {
		return
	}
	m, err := lab.Collect(sc)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// docker stats mede em % de um núcleo; mostramos em % do limite do lab
	if c := sc.Resources.CPUs; c > 0 {
		m.CPUPercent = min(m.CPUPercent/c, 100)
	}
	writeJSON(w, http.StatusOK, m)
}

func (s *Server) hint(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	n, err := strconv.Atoi(r.PathValue("n"))
	hints, herr := sc.Hints()
	if err != nil || herr != nil || n < 1 || n > len(hints) {
		writeErr(w, http.StatusBadRequest, "dica inválida")
		return
	}
	rec := s.Store.Get(sc.ID)
	if n > rec.HintsUsed+1 {
		writeErr(w, http.StatusBadRequest, "revele a dica anterior primeiro")
		return
	}
	if err := s.Store.UseHint(sc.ID, n, hints[n-1].Cost); err != nil {
		log.Printf("progresso: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"n": n, "cost": hints[n-1].Cost, "text": strings.TrimSpace(hints[n-1].Text),
	})
}

func (s *Server) solution(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	rec := s.Store.Get(sc.ID)
	if rec.StartedAt != nil && !rec.SolutionViewed {
		_ = s.Store.ViewSolution(sc.ID)
	}
	sol, _ := sc.ReadFile(sc.Content.Solution)
	root, _ := sc.ReadFile(sc.Content.RootCause)
	refs, _ := sc.ReadFile(sc.Content.References)
	writeJSON(w, http.StatusOK, map[string]any{"solution": sol, "rootCause": root, "references": refs})
}

// ---------- frontend ----------

const noFrontend = `<!doctype html><meta charset="utf-8"><title>Incidente Zero</title>
<body style="font-family:system-ui;background:#0A0E13;color:#E6EDF3;padding:40px">
<h1>Interface não compilada</h1>
<p>Rode <code>make ui</code> (ou <code>cd web &amp;&amp; npm ci &amp;&amp; npm run build</code>) e recompile o <code>iz</code>.</p>
<p>A API está no ar em <code>/api/status</code>.</p>`

func (s *Server) spa() http.Handler {
	files := http.FileServerFS(s.Static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeErr(w, http.StatusNotFound, "rota não encontrada")
			return
		}
		if _, err := fs.Stat(s.Static, "index.html"); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(noFrontend))
			return
		}
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p != "" {
			if _, err := fs.Stat(s.Static, p); errors.Is(err, fs.ErrNotExist) {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		if strings.HasPrefix(p, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}
