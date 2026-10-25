package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/plantao"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// ---------- agendador ----------

// AgendarPlantao roda o relógio do Modo Plantão até o ctx acabar: sorteia e
// sobe incidentes na hora certa, escala os que ninguém reconhece e encerra o
// turno. No modo multiusuário, cuida do plantão de cada pessoa.
func (s *Server) AgendarPlantao(ctx context.Context) {
	if s.Plantao == nil && !s.multiusuario() {
		return
	}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		for _, c := range s.contas() {
			if c.Plantao != nil {
				s.tickPlantao(c)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// quem identifica a conta nos logs.
func quem(c *conta) string {
	if c.Nome == "" {
		return ""
	}
	return " [" + c.Nome + "]"
}

func (s *Server) tickPlantao(c *conta) {
	a, err := c.Plantao.Tick()
	if err != nil {
		log.Printf("plantão%s: %v", quem(c), err)
	}
	if a.Escalou != "" {
		_ = c.Store.Log(a.Escalou, "plantao", "Escalado: alerta sem reconhecimento em 5 minutos")
		log.Printf("plantão%s: incidente %s escalado", quem(c), a.Escalou)
	}
	if a.Encerrado {
		log.Printf("plantão%s: turno encerrado", quem(c))
	}
	if a.Disparar {
		s.dispararIncidente(c)
	}
}

// candidatos devolve os cenários das trilhas escolhidas no turno.
func (s *Server) candidatos(trilhas []string) []string {
	all, err := scenario.LoadAll(s.ScenariosDir)
	if err != nil {
		return nil
	}
	quer := map[string]bool{}
	for _, t := range trilhas {
		quer[t] = true
	}
	var out []string
	for _, sc := range all {
		if quer[sc.Level] {
			out = append(out, sc.ID)
		}
	}
	return out
}

func (s *Server) dispararIncidente(c *conta) {
	turno := c.Plantao.Atual()
	if !turno.Ativo() {
		return
	}
	id := c.Plantao.Sortear(s.candidatos(turno.Trilhas))
	if id == "" {
		c.Plantao.SemCenario()
		return
	}
	sc, err := scenario.Find(s.ScenariosDir, id)
	if err != nil {
		c.Plantao.SemCenario()
		return
	}
	sc.Usuario = c.Nome
	if err := c.Plantao.Preparando(id); err != nil {
		log.Printf("plantão%s: %v", quem(c), err)
		return
	}
	log.Printf("plantão%s: preparando o incidente %s", quem(c), id)
	go func() {
		out, err := s.subirLab(sc)
		if err != nil {
			log.Printf("plantão%s: o lab %s não subiu: %v\n%s", quem(c), id, err, ultimasLinhas(out, 15))
			_ = c.Plantao.FalhouAoPreparar(id, err.Error())
			return
		}
		if err := c.Store.Start(id); err != nil {
			log.Printf("progresso: %v", err)
		}
		_ = c.Store.Log(id, "plantao", "Alerta disparado no plantão")
		s.touch(sc)
		if err := c.Plantao.Alertar(id); err != nil {
			log.Printf("plantão%s: %v", quem(c), err)
		}
		log.Printf("plantão%s: alerta do incidente %s disparado", quem(c), id)
	}()
}

// ---------- API ----------

type incidenteDTO struct {
	plantao.Incidente
	Number     string `json:"number"`
	Title      string `json:"title"`
	AlertTitle string `json:"alertTitle"`
	Severity   string `json:"severity"`
	Level      string `json:"level"`
	// Para estudar: só nos incidentes encerrados em que a pessoa travou
	Estudo *estudoDTO `json:"estudo,omitempty"`
}

type turnoDTO struct {
	ID          string         `json:"id"`
	Inicio      int64          `json:"inicio"`
	Fim         int64          `json:"fim"`
	Ritmo       string         `json:"ritmo"`
	Trilhas     []string       `json:"trilhas"`
	Encerrado   bool           `json:"encerrado"`
	EncerradoEm int64          `json:"encerradoEm,omitempty"`
	Incidentes  []incidenteDTO `json:"incidentes"`
	Resumo      plantao.Resumo `json:"resumo"`
}

func (s *Server) incidenteDTO(i plantao.Incidente) incidenteDTO {
	d := incidenteDTO{Incidente: i, Number: number(i.ID)}
	if sc, err := scenario.Find(s.ScenariosDir, i.ID); err == nil {
		d.Title, d.AlertTitle, d.Severity, d.Level = sc.Title, sc.Alert.Title, sc.Alert.Severity, sc.Level
		d.Estudo = estudoDe(i, sc)
	}
	return d
}

func (s *Server) turnoDTO(t *plantao.Turno) *turnoDTO {
	if t == nil {
		return nil
	}
	d := &turnoDTO{
		ID: t.ID, Inicio: t.Inicio.UnixMilli(), Fim: t.Fim.UnixMilli(), Ritmo: t.Ritmo,
		Trilhas: t.Trilhas, Encerrado: !t.Ativo(), Resumo: t.Resumir(),
		Incidentes: make([]incidenteDTO, 0, len(t.Incidentes)),
	}
	if t.EncerradoEm != nil {
		d.EncerradoEm = t.EncerradoEm.UnixMilli()
	}
	for _, i := range t.Incidentes {
		d.Incidentes = append(d.Incidentes, s.incidenteDTO(i))
	}
	return d
}

func (s *Server) plantaoGet(w http.ResponseWriter, r *http.Request) {
	c := contaReq(r)
	if c.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	atual := c.Plantao.Atual()
	resp := map[string]any{
		"agora":           time.Now().UnixMilli(),
		"prazoReconhecer": int(plantao.PrazoReconhecer.Seconds()),
		"turno":           nil,
		"incidente":       nil,
	}
	if atual.Ativo() {
		resp["turno"] = s.turnoDTO(atual)
		if i := atual.Atual(); i != nil {
			resp["incidente"] = s.incidenteDTO(*i)
		}
	}
	var anteriores []*turnoDTO
	for i, t := range c.Plantao.Anteriores() {
		if i == 10 {
			break
		}
		t := t
		anteriores = append(anteriores, s.turnoDTO(&t))
	}
	resp["anteriores"] = anteriores
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) plantaoAssumir(w http.ResponseWriter, r *http.Request) {
	c := contaReq(r)
	if c.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	var in struct {
		DuracaoMin int      `json:"duracaoMin"`
		Ritmo      string   `json:"ritmo"`
		Trilhas    []string `json:"trilhas"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "pedido inválido")
		return
	}
	if len(s.candidatos(in.Trilhas)) == 0 {
		writeErr(w, http.StatusBadRequest, "nenhum cenário nas trilhas escolhidas")
		return
	}
	t, err := c.Plantao.Assumir(time.Duration(in.DuracaoMin)*time.Minute, in.Ritmo, in.Trilhas)
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("plantão%s: turno assumido (%d min, ritmo %s, trilhas %v)", quem(c), in.DuracaoMin, in.Ritmo, in.Trilhas)
	writeJSON(w, http.StatusOK, map[string]any{"turno": s.turnoDTO(t)})
}

func (s *Server) plantaoEncerrar(w http.ResponseWriter, r *http.Request) {
	c := contaReq(r)
	if c.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	t, err := c.Plantao.Encerrar()
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("plantão%s: turno encerrado pela pessoa", quem(c))
	writeJSON(w, http.StatusOK, map[string]any{"turno": s.turnoDTO(t)})
}

func (s *Server) plantaoReconhecer(w http.ResponseWriter, r *http.Request) {
	c := contaReq(r)
	if c.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	i, err := c.Plantao.Reconhecer()
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	seg := int(i.ReconhecidoEm.Sub(*i.AlertaEm).Seconds())
	_ = c.Store.Log(i.ID, "plantao", "Incidente reconhecido "+duracao(seg)+" depois do alerta")
	if sc, err := scenario.Find(s.ScenariosDir, i.ID); err == nil {
		sc.Usuario = c.Nome
		s.touch(sc)
	}
	writeJSON(w, http.StatusOK, map[string]any{"incidente": s.incidenteDTO(*i)})
}

// incidenteDoPlantao informa se o cenário é o incidente aberto do plantão da conta.
func incidenteDoPlantao(c *conta, id string) bool {
	return c != nil && c.Plantao != nil && c.Plantao.IncidenteAtual() == id
}

// bloqueadoPeloPlantao impede abrir outro lab enquanto há um incidente de plantão aberto.
func bloqueadoPeloPlantao(c *conta, id string) string {
	if c == nil || c.Plantao == nil {
		return ""
	}
	atual := c.Plantao.IncidenteAtual()
	if atual == "" || atual == id {
		return ""
	}
	return "você está de plantão com o incidente " + number(atual) + " aberto: resolva-o ou encerre o turno antes de abrir outro lab"
}

// ultimasLinhas devolve as últimas n linhas de uma saída (para o log).
func ultimasLinhas(out string, n int) string {
	linhas := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(linhas) > n {
		linhas = linhas[len(linhas)-n:]
	}
	return strings.Join(linhas, "\n")
}
