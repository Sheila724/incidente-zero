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
// sobe incidentes na hora certa, escala os que ninguém reconhece e encerra o turno.
func (s *Server) AgendarPlantao(ctx context.Context) {
	if s.Plantao == nil {
		return
	}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		s.tickPlantao()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Server) tickPlantao() {
	a, err := s.Plantao.Tick()
	if err != nil {
		log.Printf("plantão: %v", err)
	}
	if a.Escalou != "" {
		_ = s.Store.Log(a.Escalou, "plantao", "Escalado: alerta sem reconhecimento em 5 minutos")
		log.Printf("plantão: incidente %s escalado", a.Escalou)
	}
	if a.Encerrado {
		log.Printf("plantão: turno encerrado")
	}
	if a.Disparar {
		s.dispararIncidente()
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

func (s *Server) dispararIncidente() {
	turno := s.Plantao.Atual()
	if !turno.Ativo() {
		return
	}
	id := s.Plantao.Sortear(s.candidatos(turno.Trilhas))
	if id == "" {
		s.Plantao.SemCenario()
		return
	}
	sc, err := scenario.Find(s.ScenariosDir, id)
	if err != nil {
		s.Plantao.SemCenario()
		return
	}
	if err := s.Plantao.Preparando(id); err != nil {
		log.Printf("plantão: %v", err)
		return
	}
	log.Printf("plantão: preparando o incidente %s", id)
	go func() {
		out, err := s.subirLab(sc)
		if err != nil {
			log.Printf("plantão: o lab %s não subiu: %v\n%s", id, err, ultimasLinhas(out, 15))
			_ = s.Plantao.FalhouAoPreparar(id, err.Error())
			return
		}
		if err := s.Store.Start(id); err != nil {
			log.Printf("progresso: %v", err)
		}
		_ = s.Store.Log(id, "plantao", "Alerta disparado no plantão")
		s.touch(id)
		if err := s.Plantao.Alertar(id); err != nil {
			log.Printf("plantão: %v", err)
		}
		log.Printf("plantão: alerta do incidente %s disparado", id)
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
	if s.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	atual := s.Plantao.Atual()
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
	for i, t := range s.Plantao.Anteriores() {
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
	if s.Plantao == nil {
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
	t, err := s.Plantao.Assumir(time.Duration(in.DuracaoMin)*time.Minute, in.Ritmo, in.Trilhas)
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("plantão: turno assumido (%d min, ritmo %s, trilhas %v)", in.DuracaoMin, in.Ritmo, in.Trilhas)
	writeJSON(w, http.StatusOK, map[string]any{"turno": s.turnoDTO(t)})
}

func (s *Server) plantaoEncerrar(w http.ResponseWriter, r *http.Request) {
	if s.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	t, err := s.Plantao.Encerrar()
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("plantão: turno encerrado pela pessoa")
	writeJSON(w, http.StatusOK, map[string]any{"turno": s.turnoDTO(t)})
}

func (s *Server) plantaoReconhecer(w http.ResponseWriter, r *http.Request) {
	if s.Plantao == nil {
		writeErr(w, http.StatusNotFound, "Modo Plantão indisponível")
		return
	}
	i, err := s.Plantao.Reconhecer()
	if err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	seg := int(i.ReconhecidoEm.Sub(*i.AlertaEm).Seconds())
	_ = s.Store.Log(i.ID, "plantao", "Incidente reconhecido "+duracao(seg)+" depois do alerta")
	s.touch(i.ID)
	writeJSON(w, http.StatusOK, map[string]any{"incidente": s.incidenteDTO(*i)})
}

// incidenteDoPlantao informa se o cenário é o incidente aberto do plantão.
func (s *Server) incidenteDoPlantao(id string) bool {
	return s.Plantao != nil && s.Plantao.IncidenteAtual() == id
}

// bloqueadoPeloPlantao impede abrir outro lab enquanto há um incidente de plantão aberto.
func (s *Server) bloqueadoPeloPlantao(id string) string {
	if s.Plantao == nil {
		return ""
	}
	atual := s.Plantao.IncidenteAtual()
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
