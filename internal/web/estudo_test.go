package web

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sheila724/incidente-zero/internal/plantao"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

func TestLinksDeReferencias(t *testing.T) {
	md := "# Para estudar mais\n\n## Linux\n\n" +
		"- [df e du](https://man7.org/df.1.html): espaço em disco\n" +
		"- [logrotate](https://example.com/logrotate) — rotação, veja [aqui](https://x.y)\n" +
		"- Cenários relacionados: **002**\n" +
		"* [Sem descrição](http://example.org/a)\n"
	ls := linksDeReferencias(md, 10)
	if len(ls) != 3 {
		t.Fatalf("esperava 3 links, veio %d: %+v", len(ls), ls)
	}
	if ls[0].Titulo != "df e du" || ls[0].URL != "https://man7.org/df.1.html" || ls[0].Desc != "espaço em disco" {
		t.Errorf("link 0 errado: %+v", ls[0])
	}
	if ls[1].Desc != "rotação, veja aqui" {
		t.Errorf("descrição com link não foi limpa: %q", ls[1].Desc)
	}
	if got := linksDeReferencias(md, 2); len(got) != 2 {
		t.Errorf("limite não respeitado: %d", len(got))
	}
}

// Todo cenário precisa render links de estudo para o relatório do plantão.
func TestTodoCenarioTemLinksDeEstudo(t *testing.T) {
	all, err := scenario.LoadAll(filepath.Join("..", "..", "scenarios"))
	if err != nil {
		t.Fatal(err)
	}
	for _, sc := range all {
		md, err := sc.ReadFile(sc.Content.References)
		if err != nil {
			t.Errorf("%s: %v", sc.ID, err)
			continue
		}
		if n := len(linksDeReferencias(md, MaxLinksEstudo)); n == 0 {
			t.Errorf("%s: referencias.md sem links em lista", sc.ID)
		}
	}
}

func TestMotivosTravou(t *testing.T) {
	sc := &scenario.Scenario{EstimatedMinutes: 20}
	alerta := time.Date(2026, 10, 23, 22, 0, 0, 0, time.UTC)
	rapido := alerta.Add(9 * time.Minute)
	lento := alerta.Add(31 * time.Minute)

	ok := plantao.Incidente{Status: plantao.Resolvido, AlertaEm: &alerta, ResolvidoEm: &rapido}
	if m := motivosTravou(ok, sc); len(m) != 0 {
		t.Errorf("resolvido rápido e sem dicas não travou: %v", m)
	}
	if estudoDe(ok, sc) != nil {
		t.Error("sem motivos não deveria haver estudo")
	}

	dif := plantao.Incidente{Status: plantao.Resolvido, AlertaEm: &alerta, ResolvidoEm: &lento, Dicas: 2, Escalado: true, SolucaoVista: true}
	m := strings.Join(motivosTravou(dif, sc), " | ")
	for _, quer := range []string{"usou 2 dicas", "escalado", "solução", "acima dos ~20 min"} {
		if !strings.Contains(m, quer) {
			t.Errorf("faltou %q em %q", quer, m)
		}
	}

	nr := plantao.Incidente{Status: plantao.NaoResolvido, AlertaEm: &alerta}
	if m := motivosTravou(nr, sc); len(m) != 1 || !strings.Contains(m[0], "não foi resolvido") {
		t.Errorf("não resolvido: %v", m)
	}
	andando := plantao.Incidente{Status: plantao.Reconhecido, AlertaEm: &alerta, Dicas: 3}
	if m := motivosTravou(andando, sc); m != nil {
		t.Errorf("incidente em andamento não entra no relatório: %v", m)
	}
}
