package web

import (
	"context"
	"log"
	"time"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// touch registra atividade num cenário. Enquanto a página do lab está aberta,
// o monitoramento consulta a API a cada poucos segundos e mantém o lab vivo.
func (s *Server) touch(id string) {
	s.seenMu.Lock()
	defer s.seenMu.Unlock()
	if s.lastSeen == nil {
		s.lastSeen = map[string]time.Time{}
	}
	s.lastSeen[id] = time.Now()
}

// idleLabs devolve os labs parados há mais de idle. Um lab que o servidor
// ainda não conhecia (iniciado pelo terminal ou antes de um restart) começa
// a contar a partir de agora. Registros de labs que não existem mais são limpos.
func idleLabs(running []string, seen map[string]time.Time, now time.Time, idle time.Duration) []string {
	alive := make(map[string]bool, len(running))
	var out []string
	for _, id := range running {
		alive[id] = true
		last, ok := seen[id]
		if !ok {
			seen[id] = now
			continue
		}
		if now.Sub(last) > idle {
			out = append(out, id)
		}
	}
	for id := range seen {
		if !alive[id] {
			delete(seen, id)
		}
	}
	return out
}

// ReapIdle destrói os labs sem atividade há mais de idle, até o ctx acabar.
// Evita que labs esquecidos fiquem consumindo CPU e memória num servidor.
func (s *Server) ReapIdle(ctx context.Context, idle time.Duration) {
	every := time.Minute
	if idle < 4*every {
		every = idle / 4
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.reapOnce(idle)
		}
	}
}

func (s *Server) reapOnce(idle time.Duration) {
	running, err := lab.Running()
	if err != nil {
		return
	}
	s.seenMu.Lock()
	if s.lastSeen == nil {
		s.lastSeen = map[string]time.Time{}
	}
	expired := idleLabs(running, s.lastSeen, time.Now(), idle)
	s.seenMu.Unlock()

	for _, id := range expired {
		if s.incidenteDoPlantao(id) {
			continue // incidente de plantão aberto: o alerta pode estar tocando sem ninguém na tela
		}
		sc, err := scenario.Find(s.ScenariosDir, id)
		if err != nil {
			continue
		}
		s.labMu.Lock()
		err = lab.Stop(sc)
		s.labMu.Unlock()
		if err != nil {
			log.Printf("limpeza: não consegui remover o lab %s: %v", id, err)
			continue
		}
		s.seenMu.Lock()
		delete(s.lastSeen, id)
		s.seenMu.Unlock()
		log.Printf("limpeza: lab %s removido após %s sem atividade", id, idle)
	}
}
