package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/progress"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// postmortemGet devolve a linha do tempo da última resolução e o que o aluno escreveu.
func (s *Server) postmortemGet(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, s.postmortemDTO(sc, s.Store.Get(sc.ID), fuso(r)))
}

// postmortemSave guarda o postmortem.
func (s *Server) postmortemSave(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	if !s.Store.Get(sc.ID).Resolved {
		writeErr(w, http.StatusConflict, "resolva o incidente antes de escrever o postmortem")
		return
	}
	var pm progress.Postmortem
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&pm); err != nil {
		writeErr(w, http.StatusBadRequest, "postmortem inválido")
		return
	}
	for _, f := range []*string{&pm.Resumo, &pm.Causa, &pm.Evidencia, &pm.Coincidencia, &pm.Acoes} {
		*f = strings.TrimSpace(*f)
	}
	rec, err := s.Store.SavePostmortem(sc.ID, pm)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.postmortemDTO(sc, rec, fuso(r)))
}

// postmortemMarkdown baixa o postmortem em Markdown.
func (s *Server) postmortemMarkdown(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.find(w, r)
	if !ok {
		return
	}
	md := PostmortemMarkdown(sc, s.Store.Get(sc.ID), fuso(r))
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="postmortem-%s.md"`, sc.ID))
	_, _ = w.Write([]byte(md))
}

// fuso usa o fuso horário do navegador (?tz=America/Sao_Paulo), ou o do servidor.
func fuso(r *http.Request) *time.Location {
	if tz := r.URL.Query().Get("tz"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	return time.Local
}

func (s *Server) postmortemDTO(sc *scenario.Scenario, rec progress.Record, loc *time.Location) map[string]any {
	pm := rec.Postmortem
	if pm == nil {
		pm = &progress.Postmortem{}
	}
	var resolvedAt *int64
	if rec.LastResolvedAt != nil {
		ms := rec.LastResolvedAt.UnixMilli()
		resolvedAt = &ms
	}
	tl := rec.LastTimeline
	if tl == nil {
		tl = []progress.Event{}
	}
	return map[string]any{
		"scenario":   s.summarize(sc, runningSet()),
		"hostname":   sc.Hostname,
		"resolved":   rec.Resolved,
		"resolvedAt": resolvedAt,
		"seconds":    rec.LastSeconds,
		"hints":      rec.LastHints,
		"points":     rec.BestPoints,
		"timeline":   tl,
		"postmortem": pm,
		"saved":      !rec.Postmortem.Vazio(),
		"markdown":   PostmortemMarkdown(sc, rec, loc),
	}
}

// PostmortemMarkdown monta o documento no formato de um postmortem sem culpados.
func PostmortemMarkdown(sc *scenario.Scenario, rec progress.Record, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	pm := rec.Postmortem
	if pm == nil {
		pm = &progress.Postmortem{}
	}
	ou := func(v, padrao string) string {
		if strings.TrimSpace(v) == "" {
			return "_" + padrao + "_"
		}
		return v
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Postmortem: %s\n\n", sc.Title)
	fmt.Fprintf(&b, "- **Alerta:** %s\n", sc.Alert.Title)
	fmt.Fprintf(&b, "- **Severidade:** %s\n", severidade(sc.Alert.Severity))
	fmt.Fprintf(&b, "- **Servidor:** `%s`\n", sc.Hostname)
	if rec.LastResolvedAt != nil {
		fmt.Fprintf(&b, "- **Resolvido em:** %s\n", rec.LastResolvedAt.In(loc).Format("02/01/2006 15:04 (MST)"))
	}
	fmt.Fprintf(&b, "- **Tempo até resolver (MTTR):** %s\n", duracao(rec.LastSeconds))
	fmt.Fprintf(&b, "- **Dicas usadas:** %d\n\n", rec.LastHints)

	fmt.Fprintf(&b, "## Resumo\n\n%s\n\n", ou(pm.Resumo, "O que aconteceu, qual foi o impacto e como foi resolvido."))

	b.WriteString("## Linha do tempo\n\n")
	if len(rec.LastTimeline) == 0 {
		b.WriteString("_Sem eventos registrados._\n\n")
	} else {
		inicio := rec.LastTimeline[0].At
		b.WriteString("| Horário | Desde o alerta | Evento |\n|---|---|---|\n")
		for _, e := range rec.LastTimeline {
			fmt.Fprintf(&b, "| %s | +%s | %s |\n", e.At.In(loc).Format("15:04:05"),
				relogio(e.At.Sub(inicio)), descricaoEvento(e, sc))
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "## Causa raiz\n\n%s\n\n", ou(pm.Causa, "Não preenchido."))
	fmt.Fprintf(&b, "## Evidência que confirmou a causa\n\n%s\n\n", ou(pm.Evidencia, "Não preenchido."))
	fmt.Fprintf(&b, "## O que só coincidiu com a recuperação\n\n%s\n\n", ou(pm.Coincidencia, "Não preenchido."))

	b.WriteString("## Ações de prevenção\n\n")
	acoes := 0
	for _, l := range strings.Split(pm.Acoes, "\n") {
		l = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(l), "-*• "))
		if l != "" {
			fmt.Fprintf(&b, "- [ ] %s\n", l)
			acoes++
		}
	}
	if acoes == 0 {
		b.WriteString("_Não preenchido._\n")
	}
	b.WriteString("\n---\n\n_Postmortem sem culpados, escrito no [Incidente Zero](https://github.com/Sheila724/incidente-zero)._\n")
	return b.String()
}

func descricaoEvento(e progress.Event, sc *scenario.Scenario) string {
	switch e.Kind {
	case "start":
		return "Plantão assumido: " + sc.Alert.Title
	default:
		return e.Detail
	}
}

func duracao(segundos int) string {
	d := time.Duration(segundos) * time.Second
	if d >= time.Hour {
		return fmt.Sprintf("%dh %02dmin", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dmin %02ds", int(d.Minutes()), int(d.Seconds())%60)
}

func relogio(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	s := int(d.Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s%3600/60, s%60)
	}
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

func severidade(s string) string {
	nomes := map[string]string{"baixa": "Baixa", "media": "Média", "alta": "Alta", "critica": "Crítica"}
	if n, ok := nomes[s]; ok {
		return n
	}
	return s
}
