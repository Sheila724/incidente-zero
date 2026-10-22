// Package plantao implementa o Modo Plantão: um turno em que incidentes caem
// em horários sorteados, sem aviso, e precisam ser reconhecidos e resolvidos.
//
// O pacote só cuida do estado e das regras (quando o próximo incidente cai,
// escalonamento, pontuação, relatório). Subir o lab e avisar a pessoa é com
// quem o usa (internal/web).
package plantao

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Estados de um incidente do plantão.
const (
	Preparando   = "preparando"  // o lab está subindo e a falha sendo injetada
	Alertando    = "alertando"   // o alerta disparou e ninguém reconheceu ainda
	Reconhecido  = "reconhecido" // alguém assumiu o incidente
	Resolvido    = "resolvido"
	NaoResolvido = "nao-resolvido" // o turno acabou com o incidente aberto
	Falhou       = "falhou"        // o lab não subiu (problema do ambiente, não da pessoa)
)

// Tempo sem reconhecimento até o incidente ser escalado.
const PrazoReconhecer = 5 * time.Minute

// Tempo extra para terminar um incidente reconhecido quando o turno acaba.
const MaxPassagem = 30 * time.Minute

// Penalidade por escalonamento.
const PenalidadeEscalado = 20

// Durações de turno aceitas.
var Duracoes = []time.Duration{30 * time.Minute, time.Hour, 2 * time.Hour, 4 * time.Hour}

// Ritmo controla o intervalo entre um incidente e o próximo.
type Ritmo struct {
	Nome            string
	Primeiro        [2]time.Duration // depois de assumir o turno
	EntreIncidentes [2]time.Duration // depois de resolver um incidente
}

var Ritmos = map[string]Ritmo{
	"calmo":   {"calmo", [2]time.Duration{4 * time.Minute, 10 * time.Minute}, [2]time.Duration{20 * time.Minute, 35 * time.Minute}},
	"normal":  {"normal", [2]time.Duration{2 * time.Minute, 6 * time.Minute}, [2]time.Duration{8 * time.Minute, 18 * time.Minute}},
	"intenso": {"intenso", [2]time.Duration{1 * time.Minute, 3 * time.Minute}, [2]time.Duration{3 * time.Minute, 7 * time.Minute}},
}

// Incidente é um incidente que caiu durante o turno.
type Incidente struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	CriadoEm      time.Time  `json:"criadoEm"`
	AlertaEm      *time.Time `json:"alertaEm,omitempty"`
	ReconhecidoEm *time.Time `json:"reconhecidoEm,omitempty"`
	ResolvidoEm   *time.Time `json:"resolvidoEm,omitempty"`
	Escalado      bool       `json:"escalado"`
	Dicas         int        `json:"dicas"`
	Pontos        int        `json:"pontos"`
	Erro          string     `json:"erro,omitempty"`
}

// Aberto informa se o incidente ainda precisa de atenção.
func (i *Incidente) Aberto() bool {
	return i.Status == Preparando || i.Status == Alertando || i.Status == Reconhecido
}

// Turno é um plantão.
type Turno struct {
	ID          string      `json:"id"`
	Inicio      time.Time   `json:"inicio"`
	Fim         time.Time   `json:"fim"`
	Ritmo       string      `json:"ritmo"`
	Trilhas     []string    `json:"trilhas"`
	EncerradoEm *time.Time  `json:"encerradoEm,omitempty"`
	Incidentes  []Incidente `json:"incidentes"`
	// Quando o próximo incidente cai. Nunca é exposto pela API: plantão não avisa.
	Proximo *time.Time `json:"proximo,omitempty"`
}

// Ativo informa se o turno ainda está valendo.
func (t *Turno) Ativo() bool { return t != nil && t.EncerradoEm == nil }

// Atual devolve o incidente aberto, se houver.
func (t *Turno) Atual() *Incidente {
	if t == nil {
		return nil
	}
	for i := range t.Incidentes {
		if t.Incidentes[i].Aberto() {
			return &t.Incidentes[i]
		}
	}
	return nil
}

// Resumo é o relatório do turno.
type Resumo struct {
	Incidentes    int `json:"incidentes"`
	Resolvidos    int `json:"resolvidos"`
	Escalados     int `json:"escalados"`
	NaoResolvidos int `json:"naoResolvidos"`
	MTTASegundos  int `json:"mttaSegundos"` // tempo médio até reconhecer
	MTTRSegundos  int `json:"mttrSegundos"` // tempo médio até resolver (desde o alerta)
	Pontos        int `json:"pontos"`
}

// Resumir calcula o relatório do turno.
func (t *Turno) Resumir() Resumo {
	var r Resumo
	var somaA, nA, somaR, nR int
	for _, i := range t.Incidentes {
		if i.Status == Falhou || i.AlertaEm == nil {
			continue
		}
		r.Incidentes++
		if i.Escalado {
			r.Escalados++
		}
		if i.ReconhecidoEm != nil {
			somaA += int(i.ReconhecidoEm.Sub(*i.AlertaEm).Seconds())
			nA++
		}
		switch i.Status {
		case Resolvido:
			r.Resolvidos++
			somaR += int(i.ResolvidoEm.Sub(*i.AlertaEm).Seconds())
			nR++
		case NaoResolvido:
			r.NaoResolvidos++
		}
		r.Pontos += i.Pontos
	}
	if nA > 0 {
		r.MTTASegundos = somaA / nA
	}
	if nR > 0 {
		r.MTTRSegundos = somaR / nR
	}
	return r
}

// Estado é o que fica gravado em disco: o turno atual e os anteriores.
type Estado struct {
	Atual      *Turno  `json:"atual,omitempty"`
	Anteriores []Turno `json:"anteriores,omitempty"`
}

// MaxAnteriores limita o histórico de turnos guardado.
const MaxAnteriores = 20

// Plantao guarda o estado com trava e persistência.
type Plantao struct {
	mu    sync.Mutex
	path  string
	est   Estado
	agora func() time.Time
	rnd   *rand.Rand
}

// DefaultPath: plantao.json ao lado do arquivo de progresso.
func DefaultPath(progressPath string) string {
	if p := os.Getenv("IZ_PLANTAO"); p != "" {
		return p
	}
	return filepath.Join(filepath.Dir(progressPath), "plantao.json")
}

// Open carrega o estado (vazio se o arquivo não existir).
func Open(path string) (*Plantao, error) {
	p := &Plantao{path: path, agora: time.Now, rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &p.est); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// SetRelogio troca o relógio e o sorteio (testes).
func (p *Plantao) SetRelogio(agora func() time.Time, seed int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.agora = agora
	p.rnd = rand.New(rand.NewSource(seed))
}

func (p *Plantao) save() error {
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p.est, "", "  ")
	if err != nil {
		return err
	}
	tmp := p.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p.path)
}

func (p *Plantao) sortear(faixa [2]time.Duration) time.Duration {
	d := faixa[1] - faixa[0]
	if d <= 0 {
		return faixa[0]
	}
	return faixa[0] + time.Duration(p.rnd.Int63n(int64(d)))
}

// copia devolve uma cópia profunda do turno.
func copia(t *Turno) *Turno {
	if t == nil {
		return nil
	}
	c := *t
	c.Trilhas = append([]string(nil), t.Trilhas...)
	c.Incidentes = append([]Incidente(nil), t.Incidentes...)
	return &c
}

// Atual devolve uma cópia do turno atual (ou nil).
func (p *Plantao) Atual() *Turno {
	p.mu.Lock()
	defer p.mu.Unlock()
	return copia(p.est.Atual)
}

// Anteriores devolve os turnos encerrados, do mais recente para o mais antigo.
func (p *Plantao) Anteriores() []Turno {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Turno, 0, len(p.est.Anteriores))
	for i := len(p.est.Anteriores) - 1; i >= 0; i-- {
		out = append(out, *copia(&p.est.Anteriores[i]))
	}
	return out
}

// Assumir começa um turno.
func (p *Plantao) Assumir(duracao time.Duration, ritmo string, trilhas []string) (*Turno, error) {
	r, ok := Ritmos[ritmo]
	if !ok {
		return nil, fmt.Errorf("ritmo inválido: %q (use calmo, normal ou intenso)", ritmo)
	}
	valida := false
	for _, d := range Duracoes {
		if d == duracao {
			valida = true
		}
	}
	if !valida {
		return nil, fmt.Errorf("duração inválida: %s", duracao)
	}
	if len(trilhas) == 0 {
		return nil, errors.New("escolha pelo menos uma trilha")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.est.Atual.Ativo() {
		return nil, errors.New("já existe um plantão em andamento")
	}
	agora := p.agora()
	prox := agora.Add(p.sortear(r.Primeiro))
	t := &Turno{
		ID: agora.Format("20060102-150405"), Inicio: agora, Fim: agora.Add(duracao),
		Ritmo: ritmo, Trilhas: trilhas, Proximo: &prox, Incidentes: []Incidente{},
	}
	p.est.Atual = t
	return copia(t), p.save()
}

// encerrar fecha o turno atual (com p.mu travado).
func (p *Plantao) encerrar(agora time.Time) {
	t := p.est.Atual
	if !t.Ativo() {
		return
	}
	for i := range t.Incidentes {
		if t.Incidentes[i].Aberto() {
			t.Incidentes[i].Status = NaoResolvido
			if t.Incidentes[i].Escalado {
				t.Incidentes[i].Pontos = -PenalidadeEscalado
			}
		}
	}
	t.EncerradoEm = &agora
	t.Proximo = nil
	p.est.Anteriores = append(p.est.Anteriores, *t)
	if n := len(p.est.Anteriores); n > MaxAnteriores {
		p.est.Anteriores = p.est.Anteriores[n-MaxAnteriores:]
	}
}

// Encerrar fecha o turno antes da hora. Devolve o turno encerrado.
func (p *Plantao) Encerrar() (*Turno, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.est.Atual.Ativo() {
		return nil, errors.New("nenhum plantão em andamento")
	}
	p.encerrar(p.agora())
	t := copia(p.est.Atual)
	return t, p.save()
}

// Acao é o que o agendador precisa fazer depois de um Tick.
type Acao struct {
	Disparar  bool   // sortear e subir um incidente agora
	Encerrado bool   // o turno acabou neste tick
	Escalou   string // id do incidente que acabou de ser escalado
}

// Tick avança o relógio do plantão: encerra o turno no fim, escala incidentes
// sem reconhecimento e avisa quando é hora de um novo incidente.
func (p *Plantao) Tick() (Acao, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var a Acao
	t := p.est.Atual
	if !t.Ativo() {
		return a, nil
	}
	agora := p.agora()
	mudou := false
	atual := t.Atual()

	if atual != nil && atual.Status == Alertando && !atual.Escalado && atual.AlertaEm != nil &&
		agora.Sub(*atual.AlertaEm) >= PrazoReconhecer {
		atual.Escalado = true
		a.Escalou = atual.ID
		mudou = true
	}

	// No fim do turno: quem já está trabalhando num incidente termina o que
	// começou (até MaxPassagem); o que ninguém reconheceu fica sem resolver.
	if !agora.Before(t.Fim) {
		trabalhando := atual != nil && (atual.Status == Preparando || atual.Status == Reconhecido)
		if !trabalhando || !agora.Before(t.Fim.Add(MaxPassagem)) {
			p.encerrar(agora)
			a.Encerrado = true
			return a, p.save()
		}
	}

	if atual == nil && t.Proximo != nil && !agora.Before(*t.Proximo) && agora.Before(t.Fim) {
		a.Disparar = true
		t.Proximo = nil // o agendador chama Disparou ou NaoDisparou
	}
	if mudou {
		return a, p.save()
	}
	return a, nil
}

// Preparando registra que um incidente foi sorteado e o lab está subindo.
func (p *Plantao) Preparando(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	t := p.est.Atual
	if !t.Ativo() {
		return errors.New("nenhum plantão em andamento")
	}
	t.Incidentes = append(t.Incidentes, Incidente{ID: id, Status: Preparando, CriadoEm: p.agora()})
	return p.save()
}

// Alertar marca que o lab está pronto e o alerta disparou.
func (p *Plantao) Alertar(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	i := p.est.Atual.Atual()
	if i == nil || i.ID != id || i.Status != Preparando {
		return errors.New("incidente não está sendo preparado")
	}
	agora := p.agora()
	i.Status = Alertando
	i.AlertaEm = &agora
	return p.save()
}

// FalhouAoPreparar registra que o lab não subiu e agenda outro incidente logo.
func (p *Plantao) FalhouAoPreparar(id, erro string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	t := p.est.Atual
	i := t.Atual()
	if i == nil || i.ID != id {
		return nil
	}
	i.Status = Falhou
	i.Erro = erro
	prox := p.agora().Add(time.Minute)
	t.Proximo = &prox
	return p.save()
}

// SemCenario é chamado quando não há cenário para sortear: o turno segue sem incidentes.
func (p *Plantao) SemCenario() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if t := p.est.Atual; t.Ativo() {
		prox := p.agora().Add(5 * time.Minute)
		t.Proximo = &prox
	}
}

// Reconhecer registra que a pessoa assumiu o incidente.
func (p *Plantao) Reconhecer() (*Incidente, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	i := p.est.Atual.Atual()
	if i == nil || i.Status != Alertando {
		return nil, errors.New("nenhum incidente aguardando reconhecimento")
	}
	agora := p.agora()
	i.Status = Reconhecido
	i.ReconhecidoEm = &agora
	c := *i
	return &c, p.save()
}

// Resolver registra a resolução do incidente atual, se for este cenário.
// pontos é a pontuação da tentativa; a penalidade de escalonamento é aplicada aqui.
func (p *Plantao) Resolver(id string, pontos, dicas int) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	t := p.est.Atual
	i := t.Atual()
	if i == nil || i.ID != id || (i.Status != Alertando && i.Status != Reconhecido) {
		return false, nil
	}
	agora := p.agora()
	if i.ReconhecidoEm == nil {
		i.ReconhecidoEm = &agora
	}
	i.Status = Resolvido
	i.ResolvidoEm = &agora
	i.Dicas = dicas
	if i.Escalado {
		pontos -= PenalidadeEscalado
	}
	if pontos < 0 {
		pontos = 0
	}
	i.Pontos = pontos
	if !agora.Before(t.Fim) {
		// Incidente terminado depois do fim do turno: encerra a passagem
		p.encerrar(agora)
		return true, p.save()
	}
	r := Ritmos[t.Ritmo]
	prox := agora.Add(p.sortear(r.EntreIncidentes))
	if prox.Before(t.Fim.Add(-3 * time.Minute)) {
		t.Proximo = &prox
	} else {
		t.Proximo = nil // não dá tempo de outro incidente neste turno
	}
	return true, p.save()
}

// IncidenteAtual devolve o id do incidente aberto do turno ativo ("" se não houver).
func (p *Plantao) IncidenteAtual() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.est.Atual.Ativo() {
		return ""
	}
	if i := p.est.Atual.Atual(); i != nil {
		return i.ID
	}
	return ""
}

// Sortear escolhe um cenário entre os candidatos, evitando os que já caíram no turno.
func (p *Plantao) Sortear(candidatos []string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(candidatos) == 0 {
		return ""
	}
	vistos := map[string]bool{}
	if t := p.est.Atual; t != nil {
		for _, i := range t.Incidentes {
			vistos[i.ID] = true
		}
	}
	var novos []string
	for _, c := range candidatos {
		if !vistos[c] {
			novos = append(novos, c)
		}
	}
	if len(novos) == 0 {
		novos = candidatos
	}
	return novos[p.rnd.Intn(len(novos))]
}
