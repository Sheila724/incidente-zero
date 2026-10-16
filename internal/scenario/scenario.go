// Package scenario carrega e localiza os cenários do Incidente Zero.
package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Mount struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
	Size string `yaml:"size"`
}

type Resources struct {
	CPUs   float64 `yaml:"cpus"`
	Memory string  `yaml:"memory"`
	PIDs   int     `yaml:"pids"`
	Mounts []Mount `yaml:"mounts"`
}

type Alert struct {
	Severity string `yaml:"severity"`
	Title    string `yaml:"title"`
	Briefing string `yaml:"briefing"`
}

type Content struct {
	Concepts   string `yaml:"concepts"`
	Hints      string `yaml:"hints"`
	Solution   string `yaml:"solution"`
	RootCause  string `yaml:"root_cause"`
	References string `yaml:"references"`
}

type Scripts struct {
	Setup  string `yaml:"setup"`
	Inject string `yaml:"inject"`
	Check  string `yaml:"check"`
}

type Scenario struct {
	ID               string    `yaml:"id"`
	Title            string    `yaml:"title"`
	Level            string    `yaml:"level"`
	Track            string    `yaml:"track"`
	EstimatedMinutes int       `yaml:"estimated_minutes"`
	Image            string    `yaml:"image"`
	Hostname         string    `yaml:"hostname"`
	User             string    `yaml:"user"`
	Resources        Resources `yaml:"resources"`
	Alert            Alert     `yaml:"alert"`
	Content          Content   `yaml:"content"`
	Scripts          Scripts   `yaml:"scripts"`

	Dir string `yaml:"-"` // pasta do cenário no disco
}

type Hint struct {
	Level int    `yaml:"level"`
	Cost  int    `yaml:"cost"`
	Text  string `yaml:"text"`
}

// Path resolve um caminho relativo à pasta do cenário.
func (s *Scenario) Path(rel string) string {
	return filepath.Join(s.Dir, rel)
}

// ContainerName é o nome do container do lab.
func (s *Scenario) ContainerName() string {
	return "iz-" + s.ID
}

// ReadFile lê um arquivo de conteúdo do cenário.
func (s *Scenario) ReadFile(rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("arquivo não definido no scenario.yaml")
	}
	b, err := os.ReadFile(s.Path(rel))
	return string(b), err
}

// Hints carrega as dicas do cenário.
func (s *Scenario) Hints() ([]Hint, error) {
	raw, err := s.ReadFile(s.Content.Hints)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Hints []Hint `yaml:"hints"`
	}
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("hints.yaml inválido: %w", err)
	}
	return doc.Hints, nil
}

// LoadAll encontra todos os scenario.yaml em root/<nivel>/<cenario>/.
func LoadAll(root string) ([]*Scenario, error) {
	files, err := filepath.Glob(filepath.Join(root, "*", "*", "scenario.yaml"))
	if err != nil {
		return nil, err
	}
	var list []*Scenario
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var s Scenario
		if err := yaml.Unmarshal(b, &s); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		s.Dir = filepath.Dir(f)
		if s.User == "" {
			s.User = "aluno"
		}
		list = append(list, &s)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list, nil
}

// Find localiza um cenário pelo ID completo ou prefixo ("001" -> "001-disco-cheio").
func Find(root, id string) (*Scenario, error) {
	all, err := LoadAll(root)
	if err != nil {
		return nil, err
	}
	var matches []*Scenario
	for _, s := range all {
		if s.ID == id {
			return s, nil
		}
		if strings.HasPrefix(s.ID, id) {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("cenário %q não encontrado (use `iz list`)", id)
	default:
		return nil, fmt.Errorf("%q corresponde a mais de um cenário; seja mais específica", id)
	}
}
