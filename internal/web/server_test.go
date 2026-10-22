package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Sheila724/incidente-zero/internal/progress"
)

func TestParseCheck(t *testing.T) {
	out := "✅ Serviço ativo\nlinha qualquer\n❌ Site respondendo HTTP 503\n  ✅ Bônus: algo\n⚠️ Rollback não resolve\n⚠ Sem o seletor\n"
	items := parseCheck(out)
	if len(items) != 5 {
		t.Fatalf("esperava 5 itens, veio %d", len(items))
	}
	for _, i := range []int{3, 4} {
		if items[i].OK || !items[i].Contorno {
			t.Errorf("item %d deveria ser contorno: %+v", i, items[i])
		}
	}
	if items[3].Text != "Rollback não resolve" || items[4].Text != "Sem o seletor" {
		t.Errorf("texto do contorno errado: %q / %q", items[3].Text, items[4].Text)
	}
	if items[1].Contorno {
		t.Errorf("❌ não é contorno: %+v", items[1])
	}
	if !items[0].OK || items[0].Text != "Serviço ativo" {
		t.Errorf("item 0 errado: %+v", items[0])
	}
	if items[1].OK || items[1].Text != "Site respondendo HTTP 503" {
		t.Errorf("item 1 errado: %+v", items[1])
	}
}

func newServer(t *testing.T, static fstest.MapFS) *httptest.Server {
	t.Helper()
	store, err := progress.Open(filepath.Join(t.TempDir(), "p.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		ScenariosDir: filepath.Join("..", "..", "scenarios"),
		Store:        store,
		Static:       static,
		Version:      "teste",
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func TestDetalheEDicas(t *testing.T) {
	ts := newServer(t, fstest.MapFS{})

	res, err := http.Get(ts.URL + "/api/scenarios/001")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Scenario struct{ ID string }
		Briefing string
		Hints    []struct {
			N    int
			Text string
		}
	}
	_ = json.NewDecoder(res.Body).Decode(&d)
	res.Body.Close()
	if d.Scenario.ID != "001-disco-cheio" || d.Briefing == "" {
		t.Fatalf("detalhe incompleto: %+v", d.Scenario)
	}
	for _, h := range d.Hints {
		if h.Text != "" {
			t.Error("dicas não podem vir reveladas antes de pedir")
		}
	}

	// Pular a ordem das dicas não é permitido
	res, _ = http.Post(ts.URL+"/api/scenarios/001/hints/2", "", nil)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("dica 2 antes da 1: esperava 400, veio %d", res.StatusCode)
	}
	res.Body.Close()

	res, _ = http.Post(ts.URL+"/api/scenarios/001/hints/1", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Errorf("dica 1: esperava 200, veio %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCabecalhosERotas(t *testing.T) {
	ts := newServer(t, fstest.MapFS{"index.html": {Data: []byte("<h1>ui</h1>")}})

	res, _ := http.Get(ts.URL + "/")
	if res.Header.Get("Content-Security-Policy") == "" || res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("cabeçalhos de segurança ausentes")
	}
	res.Body.Close()

	res, _ = http.Get(ts.URL + "/api/nao-existe")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("rota inexistente da API: esperava 404, veio %d", res.StatusCode)
	}
	res.Body.Close()

	res, _ = http.Get(ts.URL + "/api/scenarios/999")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("cenário inexistente: esperava 404, veio %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestSemInterfaceCompilada(t *testing.T) {
	ts := newServer(t, fstest.MapFS{})
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if !strings.Contains(string(body), "Interface não compilada") {
		t.Error("sem dist/index.html, a página deveria explicar como compilar a interface")
	}
}
