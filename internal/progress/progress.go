// Package progress guarda o histórico de plantões do aluno num arquivo JSON local.
package progress

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Record é o estado de um cenário para este aluno.
type Record struct {
	// Tentativa atual
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	HintsUsed      int        `json:"hintsUsed"`
	HintCost       int        `json:"hintCost"`
	SolutionViewed bool       `json:"solutionViewed"`

	// Melhor resultado já obtido
	Resolved       bool       `json:"resolved"`
	LastResolvedAt *time.Time `json:"lastResolvedAt,omitempty"`
	LastSeconds    int        `json:"lastSeconds"`
	LastHints      int        `json:"lastHints"`
	BestSeconds    int        `json:"bestSeconds"`
	BestPoints     int        `json:"bestPoints"`
	Attempts       int        `json:"attempts"`

	// Postmortem
	Timeline     []Event     `json:"timeline,omitempty"`     // eventos da tentativa atual
	LastTimeline []Event     `json:"lastTimeline,omitempty"` // eventos da última resolução
	Postmortem   *Postmortem `json:"postmortem,omitempty"`
}

// Event é um momento do plantão, usado na linha do tempo do postmortem.
type Event struct {
	At     time.Time `json:"at"`
	Kind   string    `json:"kind"` // start | hint | check | solution | resolved
	Detail string    `json:"detail,omitempty"`
}

// Postmortem é o que o aluno escreveu sobre o incidente.
type Postmortem struct {
	Resumo       string    `json:"resumo"`
	Causa        string    `json:"causa"`
	Evidencia    string    `json:"evidencia"`
	Coincidencia string    `json:"coincidencia"`
	Acoes        string    `json:"acoes"`
	SavedAt      time.Time `json:"savedAt"`
}

// Vazio informa se nada foi escrito.
func (p *Postmortem) Vazio() bool {
	return p == nil || (p.Resumo == "" && p.Causa == "" && p.Evidencia == "" && p.Coincidencia == "" && p.Acoes == "")
}

type Store struct {
	mu   sync.Mutex
	path string
	data map[string]*Record
}

// DefaultPath: ~/.config/incidente-zero/progress.json (ou IZ_PROGRESS).
func DefaultPath() string {
	if p := os.Getenv("IZ_PROGRESS"); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "incidente-zero", "progress.json")
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, data: map[string]*Record{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) rec(id string) *Record {
	r, ok := s.data[id]
	if !ok {
		r = &Record{}
		s.data[id] = r
	}
	return r
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Get devolve uma cópia do registro (zero se não existir).
func (s *Store) Get(id string) Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.data[id]; ok {
		return *r
	}
	return Record{}
}

// All devolve uma cópia de todos os registros.
func (s *Store) All() map[string]Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Record, len(s.data))
	for k, v := range s.data {
		out[k] = *v
	}
	return out
}

// Start abre uma nova tentativa (zera dicas e cronômetro).
func (s *Store) Start(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	now := time.Now()
	r.StartedAt = &now
	r.HintsUsed, r.HintCost, r.SolutionViewed = 0, 0, false
	r.Attempts++
	r.Timeline = []Event{{At: now, Kind: "start"}}
	return s.save()
}

// UseHint registra a dica n (1-based) com o custo informado.
func (s *Store) UseHint(id string, n, cost int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	if n > r.HintsUsed {
		r.HintsUsed = n
		r.HintCost += cost
		r.log("hint", fmt.Sprintf("Dica %d revelada (−%d pontos)", n, cost))
	}
	return s.save()
}

func (s *Store) ViewSolution(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	if !r.SolutionViewed {
		r.log("solution", "Solução comentada aberta")
	}
	r.SolutionViewed = true
	return s.save()
}

// Resolve fecha a tentativa atual e calcula a pontuação:
// base − custo das dicas, +10 se dentro do tempo bônus, 0 se a solução foi aberta antes.
func (s *Store) Resolve(id string, base, bonusMinutes int) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	now := time.Now()
	secs := 0
	if r.StartedAt != nil {
		secs = int(now.Sub(*r.StartedAt).Seconds())
	}
	if base <= 0 {
		base = 100
	}
	points := base - r.HintCost
	if bonusMinutes > 0 && r.StartedAt != nil && secs <= bonusMinutes*60 {
		points += 10
	}
	if r.SolutionViewed {
		points = 0
	}
	if points < 0 {
		points = 0
	}

	r.Resolved = true
	r.LastResolvedAt = &now
	r.LastSeconds = secs
	r.LastHints = r.HintsUsed
	if r.BestSeconds == 0 || (secs > 0 && secs < r.BestSeconds) {
		r.BestSeconds = secs
	}
	if points > r.BestPoints {
		r.BestPoints = points
	}
	if r.StartedAt != nil {
		r.log("resolved", "Verificação aprovada: incidente resolvido")
		r.LastTimeline = r.Timeline
		r.Timeline = nil
	}
	r.StartedAt = nil
	return *r, s.save()
}

// log registra um evento na tentativa em andamento.
func (r *Record) log(kind, detail string) {
	if r.StartedAt == nil {
		return
	}
	r.Timeline = append(r.Timeline, Event{At: time.Now(), Kind: kind, Detail: detail})
}

// Log registra um evento na tentativa em andamento (ex.: uma verificação).
func (s *Store) Log(id, kind, detail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	if r.StartedAt == nil {
		return nil
	}
	r.log(kind, detail)
	return s.save()
}

// SavePostmortem guarda o postmortem do aluno para o cenário.
func (s *Store) SavePostmortem(id string, pm Postmortem) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rec(id)
	pm.SavedAt = time.Now()
	r.Postmortem = &pm
	return *r, s.save()
}

// HistoryEntry é uma linha do histórico de plantões.
type HistoryEntry struct {
	ID         string    `json:"id"`
	ResolvedAt time.Time `json:"resolvedAt"`
	Seconds    int       `json:"seconds"`
	Hints      int       `json:"hints"`
	Points     int       `json:"points"`
	Postmortem bool      `json:"postmortem"` // há postmortem escrito
}

// History lista os cenários resolvidos, do mais recente para o mais antigo.
func (s *Store) History() []HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []HistoryEntry
	for id, r := range s.data {
		if r.Resolved && r.LastResolvedAt != nil {
			out = append(out, HistoryEntry{ID: id, ResolvedAt: *r.LastResolvedAt,
				Seconds: r.LastSeconds, Hints: r.LastHints, Points: r.BestPoints,
				Postmortem: !r.Postmortem.Vazio()})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ResolvedAt.After(out[j].ResolvedAt) })
	return out
}
