package progress

import (
	"path/filepath"
	"testing"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "progress.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPontuacao(t *testing.T) {
	cases := []struct {
		nome     string
		dicas    [][2]int // {n, custo}
		solucao  bool
		bonusMin int
		esperado int
	}{
		{"sem dicas, dentro do bônus", nil, false, 10, 110},
		{"sem dicas, sem bônus configurado", nil, false, 0, 100},
		{"duas dicas", [][2]int{{1, 10}, {2, 20}}, false, 0, 70},
		{"dica repetida conta uma vez", [][2]int{{1, 10}, {1, 10}}, false, 0, 90},
		{"solução aberta zera", [][2]int{{1, 10}}, true, 10, 0},
		{"custo maior que a base não fica negativo", [][2]int{{1, 80}, {2, 80}}, false, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.nome, func(t *testing.T) {
			s := open(t)
			if err := s.Start("x"); err != nil {
				t.Fatal(err)
			}
			for _, d := range c.dicas {
				_ = s.UseHint("x", d[0], d[1])
			}
			if c.solucao {
				_ = s.ViewSolution("x")
			}
			r, err := s.Resolve("x", 100, c.bonusMin)
			if err != nil {
				t.Fatal(err)
			}
			if r.BestPoints != c.esperado {
				t.Errorf("pontos = %d, esperava %d", r.BestPoints, c.esperado)
			}
			if !r.Resolved || r.StartedAt != nil {
				t.Error("a tentativa deveria estar fechada e marcada como resolvida")
			}
		})
	}
}

func TestPersistencia(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.json")
	s, _ := Open(path)
	_ = s.Start("001")
	_, _ = s.Resolve("001", 100, 0)

	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s2.Get("001").Resolved {
		t.Error("o progresso não foi salvo em disco")
	}
	if len(s2.History()) != 1 {
		t.Error("o histórico deveria ter uma entrada")
	}
}
