package scenario

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Valida todos os cenários do repositório: o que um contribuidor esquecer, o CI pega.
func TestCenariosDoRepositorio(t *testing.T) {
	root := filepath.Join("..", "..", "scenarios")
	all, err := LoadAll(root)
	if err != nil {
		t.Fatalf("carregando cenários: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("nenhum cenário encontrado")
	}

	levels := map[string]bool{"iniciante": true, "intermediario": true, "avancado": true}
	severities := map[string]bool{"baixa": true, "media": true, "alta": true, "critica": true}
	idRe := regexp.MustCompile(`^\d{3}-[a-z0-9-]+$`)
	numbers := map[string]string{}

	for _, s := range all {
		t.Run(s.ID, func(t *testing.T) {
			if !idRe.MatchString(s.ID) {
				t.Errorf("id %q fora do padrão NNN-nome-em-minusculas", s.ID)
			}
			if base := filepath.Base(s.Dir); base != s.ID {
				t.Errorf("id %q diferente do nome da pasta %q", s.ID, base)
			}
			if lvl := filepath.Base(filepath.Dir(s.Dir)); lvl != s.Level {
				t.Errorf("level %q diferente da pasta do nível %q", s.Level, lvl)
			}
			num := strings.SplitN(s.ID, "-", 2)[0]
			if other, dup := numbers[num]; dup {
				t.Errorf("número %s repetido (também em %s)", num, other)
			}
			numbers[num] = s.ID

			if s.Title == "" || s.Image == "" || s.Hostname == "" {
				t.Error("title, image e hostname são obrigatórios")
			}
			if !levels[s.Level] {
				t.Errorf("level inválido: %q", s.Level)
			}
			if !severities[s.Alert.Severity] {
				t.Errorf("alert.severity inválida: %q", s.Alert.Severity)
			}
			if s.Alert.Title == "" {
				t.Error("alert.title é obrigatório")
			}
			if s.EstimatedMinutes <= 0 {
				t.Error("estimated_minutes deve ser maior que zero")
			}
			if len(s.Objectives) == 0 {
				t.Error("learning_objectives não pode ser vazio")
			}

			files := map[string]string{
				"alert.briefing":     s.Alert.Briefing,
				"content.concepts":   s.Content.Concepts,
				"content.hints":      s.Content.Hints,
				"content.solution":   s.Content.Solution,
				"content.root_cause": s.Content.RootCause,
				"content.references": s.Content.References,
				"scripts.setup":      s.Scripts.Setup,
				"scripts.inject":     s.Scripts.Inject,
				"scripts.check":      s.Scripts.Check,
			}
			for field, rel := range files {
				if rel == "" {
					t.Errorf("%s não definido", field)
					continue
				}
				if _, err := os.Stat(s.Path(rel)); err != nil {
					t.Errorf("%s aponta para um arquivo que não existe: %s", field, rel)
				}
			}

			hints, err := s.Hints()
			if err != nil {
				t.Fatalf("hints: %v", err)
			}
			if len(hints) == 0 {
				t.Error("o cenário precisa de pelo menos uma dica")
			}
			for i, h := range hints {
				if strings.TrimSpace(h.Text) == "" || h.Cost <= 0 {
					t.Errorf("dica %d sem texto ou sem custo", i+1)
				}
			}

			for i, g := range s.Guided {
				if g.Title == "" || g.Text == "" {
					t.Errorf("passo guiado %d sem título ou texto", i+1)
				}
			}

			for _, m := range s.Resources.Mounts {
				if m.Type != "tmpfs" || !strings.HasPrefix(m.Path, "/") {
					t.Errorf("mount inválido: %+v", m)
				}
			}

			inject, _ := os.ReadFile(s.Path(s.Scripts.Inject))
			if !strings.Contains(string(inject), "Autoverificação") {
				t.Error("o inject.sh deve terminar com uma autoverificação de que a falha apareceu")
			}
		})
	}
}

func TestFindPorPrefixo(t *testing.T) {
	root := filepath.Join("..", "..", "scenarios")
	s, err := Find(root, "001")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "001-disco-cheio" {
		t.Errorf("esperava 001-disco-cheio, veio %s", s.ID)
	}
	if _, err := Find(root, "999"); err == nil {
		t.Error("esperava erro para um cenário inexistente")
	}
}
