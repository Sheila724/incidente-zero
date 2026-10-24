package web

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/plantao"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// Material de estudo para os incidentes em que a pessoa travou no plantão.

type linkEstudo struct {
	Titulo string `json:"titulo"`
	URL    string `json:"url"`
	Desc   string `json:"desc,omitempty"`
}

type estudoDTO struct {
	Motivos []string     `json:"motivos"`
	Links   []linkEstudo `json:"links"`
}

// MaxLinksEstudo limita os links por incidente, para o relatório não virar uma lista.
const MaxLinksEstudo = 4

// Itens de lista do referencias.md: "- [Título](https://...): descrição"
var reLinkLista = regexp.MustCompile(`^\s*[-*]\s+\[([^\]]+)\]\((https?://[^)\s]+)\)\s*[:—–-]?\s*(.*)$`)

var reLinkTexto = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)

// linksDeReferencias extrai os links das listas do referencias.md do cenário.
func linksDeReferencias(md string, max int) []linkEstudo {
	var out []linkEstudo
	for _, linha := range strings.Split(md, "\n") {
		m := reLinkLista.FindStringSubmatch(linha)
		if m == nil {
			continue
		}
		desc := strings.TrimSpace(m[3])
		// Descrições com outros links markdown viram texto simples
		desc = reLinkTexto.ReplaceAllString(desc, "$1")
		out = append(out, linkEstudo{Titulo: strings.TrimSpace(m[1]), URL: m[2], Desc: desc})
		if len(out) == max {
			break
		}
	}
	return out
}

// motivosTravou lista o que indica que a pessoa teve dificuldade no incidente.
func motivosTravou(i plantao.Incidente, sc *scenario.Scenario) []string {
	var m []string
	switch i.Status {
	case plantao.NaoResolvido:
		m = append(m, "não foi resolvido no turno")
	case plantao.Resolvido:
	default:
		return nil // ainda em andamento
	}
	if i.Escalado {
		m = append(m, "foi escalado (alerta sem reconhecimento)")
	}
	if i.SolucaoVista {
		m = append(m, "a solução foi aberta")
	}
	if i.Dicas > 0 {
		if i.Dicas == 1 {
			m = append(m, "usou 1 dica")
		} else {
			m = append(m, fmt.Sprintf("usou %d dicas", i.Dicas))
		}
	}
	if i.Status == plantao.Resolvido && i.AlertaEm != nil && i.ResolvidoEm != nil && sc != nil && sc.EstimatedMinutes > 0 {
		gasto := i.ResolvidoEm.Sub(*i.AlertaEm)
		if gasto > time.Duration(sc.EstimatedMinutes)*time.Minute {
			m = append(m, fmt.Sprintf("levou %s, acima dos ~%d min estimados", duracao(int(gasto.Seconds())), sc.EstimatedMinutes))
		}
	}
	return m
}

// estudoDe monta o material de estudo de um incidente (nil se a pessoa não travou).
func estudoDe(i plantao.Incidente, sc *scenario.Scenario) *estudoDTO {
	motivos := motivosTravou(i, sc)
	if len(motivos) == 0 {
		return nil
	}
	e := &estudoDTO{Motivos: motivos, Links: []linkEstudo{}}
	if sc != nil && sc.Content.References != "" {
		if md, err := sc.ReadFile(sc.Content.References); err == nil {
			e.Links = linksDeReferencias(md, MaxLinksEstudo)
		}
	}
	return e
}
