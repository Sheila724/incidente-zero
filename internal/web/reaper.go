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
// A chave é o nome do container (um lab por usuário e cenário).
func (s *Server) touch(sc *scenario.Scenario) {
	s.seenMu.Lock()
	defer s.seenMu.Unlock()
	if s.lastSeen == nil {
		s.lastSeen = map[string]time.Time{}
	}
	s.lastSeen[sc.ContainerName()] = time.Now()
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
	labs, err := lab.RunningAll()
	if err != nil {
		return
	}
	porNome := map[string]lab.Lab{}
	nomes := make([]string, 0, len(labs))
	for _, l := range labs {
		n := (&scenario.Scenario{ID: l.ID, Usuario: l.Usuario}).ContainerName()
		porNome[n] = l
		nomes = append(nomes, n)
	}
	s.seenMu.Lock()
	if s.lastSeen == nil {
		s.lastSeen = map[string]time.Time{}
	}
	expired := idleLabs(nomes, s.lastSeen, time.Now(), idle)
	s.seenMu.Unlock()

	for _, nome := range expired {
		l := porNome[nome]
		if incidenteDoPlantao(s.contaPorNome(l.Usuario), l.ID) {
			continue // incidente de plantão aberto: o alerta pode estar tocando sem ninguém na tela
		}
		sc, err := scenario.Find(s.ScenariosDir, l.ID)
		if err != nil {
			continue
		}
		sc.Usuario = l.Usuario
		s.labMu.Lock()
		err = lab.Stop(sc)
		s.labMu.Unlock()
		if err != nil {
			log.Printf("limpeza: não consegui remover o lab %s: %v", nome, err)
			continue
		}
		s.seenMu.Lock()
		delete(s.lastSeen, nome)
		s.seenMu.Unlock()
		log.Printf("limpeza: lab %s removido após %s sem atividade", nome, idle)
	}
}
