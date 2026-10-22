package plantao

import (
	"path/filepath"
	"testing"
	"time"
)

type relogio struct{ t time.Time }

func (r *relogio) agora() time.Time       { return r.t }
func (r *relogio) passar(d time.Duration) { r.t = r.t.Add(d) }

func novo(t *testing.T) (*Plantao, *relogio) {
	t.Helper()
	p, err := Open(filepath.Join(t.TempDir(), "plantao.json"))
	if err != nil {
		t.Fatal(err)
	}
	r := &relogio{time.Date(2026, 10, 22, 2, 0, 0, 0, time.UTC)}
	p.SetRelogio(r.agora, 1)
	return p, r
}

// esperarDisparo avança o relógio até o plantão pedir um incidente.
func esperarDisparo(t *testing.T, p *Plantao, r *relogio, max time.Duration) {
	t.Helper()
	for passou := time.Duration(0); passou <= max; passou += 30 * time.Second {
		a, err := p.Tick()
		if err != nil {
			t.Fatal(err)
		}
		if a.Disparar {
			return
		}
		r.passar(30 * time.Second)
	}
	t.Fatalf("nenhum incidente em %s", max)
}

func TestValidacao(t *testing.T) {
	p, _ := novo(t)
	if _, err := p.Assumir(time.Hour, "frenético", []string{"iniciante"}); err == nil {
		t.Error("ritmo inválido aceito")
	}
	if _, err := p.Assumir(17*time.Minute, "normal", []string{"iniciante"}); err == nil {
		t.Error("duração inválida aceita")
	}
	if _, err := p.Assumir(time.Hour, "normal", nil); err == nil {
		t.Error("turno sem trilhas aceito")
	}
	if _, err := p.Assumir(time.Hour, "normal", []string{"iniciante"}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Assumir(time.Hour, "normal", []string{"iniciante"}); err == nil {
		t.Error("dois plantões ao mesmo tempo")
	}
}

func TestCicloCompleto(t *testing.T) {
	p, r := novo(t)
	if _, err := p.Assumir(time.Hour, "normal", []string{"iniciante"}); err != nil {
		t.Fatal(err)
	}
	// Nada cai antes do mínimo do ritmo
	if a, _ := p.Tick(); a.Disparar {
		t.Fatal("incidente no primeiro segundo do turno")
	}
	esperarDisparo(t, p, r, 7*time.Minute)

	id := p.Sortear([]string{"001-disco-cheio"})
	if err := p.Preparando(id); err != nil {
		t.Fatal(err)
	}
	if p.IncidenteAtual() != id {
		t.Fatal("incidente atual errado")
	}
	// Enquanto prepara, não dispara outro
	if a, _ := p.Tick(); a.Disparar {
		t.Fatal("disparou dois incidentes")
	}
	r.passar(40 * time.Second)
	if err := p.Alertar(id); err != nil {
		t.Fatal(err)
	}
	r.passar(90 * time.Second)
	if _, err := p.Reconhecer(); err != nil {
		t.Fatal(err)
	}
	r.passar(10 * time.Minute)
	ok, err := p.Resolver(id, 90, 1)
	if err != nil || !ok {
		t.Fatalf("resolver: %v %v", ok, err)
	}
	// Resolver outro cenário não mexe em nada
	if ok, _ := p.Resolver("002-nginx-nao-sobe", 100, 0); ok {
		t.Fatal("resolveu cenário que não é do plantão")
	}

	turno := p.Atual()
	res := turno.Resumir()
	if res.Incidentes != 1 || res.Resolvidos != 1 || res.MTTASegundos != 90 || res.MTTRSegundos != 690 || res.Pontos != 90 {
		t.Errorf("resumo errado: %+v", res)
	}
	if turno.Proximo == nil {
		t.Error("próximo incidente não foi agendado")
	}
}

func TestEscalonamento(t *testing.T) {
	p, r := novo(t)
	_, _ = p.Assumir(time.Hour, "intenso", []string{"iniciante"})
	esperarDisparo(t, p, r, 4*time.Minute)
	_ = p.Preparando("004-cpu-100")
	_ = p.Alertar("004-cpu-100")

	r.passar(PrazoReconhecer - time.Second)
	if a, _ := p.Tick(); a.Escalou != "" {
		t.Fatal("escalou antes do prazo")
	}
	r.passar(2 * time.Second)
	if a, _ := p.Tick(); a.Escalou != "004-cpu-100" {
		t.Fatal("não escalou no prazo")
	}
	if a, _ := p.Tick(); a.Escalou != "" {
		t.Fatal("escalou duas vezes")
	}
	_, _ = p.Resolver("004-cpu-100", 100, 0)
	i := p.Atual().Incidentes[0]
	if !i.Escalado || i.Pontos != 100-PenalidadeEscalado {
		t.Errorf("penalidade não aplicada: %+v", i)
	}
}

func TestFimDoTurno(t *testing.T) {
	p, r := novo(t)
	_, _ = p.Assumir(30*time.Minute, "intenso", []string{"iniciante"})
	esperarDisparo(t, p, r, 4*time.Minute)
	_ = p.Preparando("004-cpu-100")
	_ = p.Alertar("004-cpu-100")
	_, _ = p.Reconhecer()

	// O turno acaba, mas quem está trabalhando termina o incidente
	r.t = p.Atual().Fim.Add(time.Minute)
	if a, _ := p.Tick(); a.Encerrado || a.Disparar {
		t.Fatalf("encerrou no meio do incidente: %+v", a)
	}
	ok, _ := p.Resolver("004-cpu-100", 80, 0)
	if !ok {
		t.Fatal("não resolveu na passagem")
	}
	if p.Atual().Ativo() {
		t.Fatal("o turno deveria ter encerrado ao resolver depois do fim")
	}
	if len(p.Anteriores()) != 1 {
		t.Fatal("turno não foi para o histórico")
	}
}

func TestFimSemReconhecer(t *testing.T) {
	p, r := novo(t)
	_, _ = p.Assumir(30*time.Minute, "intenso", []string{"iniciante"})
	esperarDisparo(t, p, r, 4*time.Minute)
	_ = p.Preparando("004-cpu-100")
	_ = p.Alertar("004-cpu-100")

	r.t = p.Atual().Fim
	a, _ := p.Tick()
	if !a.Encerrado {
		t.Fatal("turno não encerrou")
	}
	i := p.Anteriores()[0].Incidentes[0]
	if i.Status != NaoResolvido || i.Pontos != -PenalidadeEscalado {
		t.Errorf("incidente ignorado deveria ficar sem resolver e escalado: %+v", i)
	}
	if _, err := p.Reconhecer(); err == nil {
		t.Error("reconheceu depois do fim do turno")
	}
}

func TestSorteioEvitaRepetir(t *testing.T) {
	p, r := novo(t)
	_, _ = p.Assumir(4*time.Hour, "intenso", []string{"iniciante"})
	cands := []string{"a", "b", "c"}
	vistos := map[string]bool{}
	for n := 0; n < 3; n++ {
		esperarDisparo(t, p, r, 10*time.Minute)
		id := p.Sortear(cands)
		if vistos[id] {
			t.Fatalf("repetiu %s antes de esgotar os candidatos", id)
		}
		vistos[id] = true
		_ = p.Preparando(id)
		_ = p.Alertar(id)
		_, _ = p.Resolver(id, 100, 0)
	}
}

func TestPersistencia(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plantao.json")
	p, _ := Open(path)
	_, _ = p.Assumir(time.Hour, "calmo", []string{"avancado"})
	q, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := q.Atual(); !a.Ativo() || a.Ritmo != "calmo" || a.Proximo == nil {
		t.Errorf("estado não foi recarregado: %+v", a)
	}
}
