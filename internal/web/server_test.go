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

	"github.com/Sheila724/incidente-zero/internal/plantao"
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

func TestPlantaoAPI(t *testing.T) {
	store, err := progress.Open(filepath.Join(t.TempDir(), "p.json"))
	if err != nil {
		t.Fatal(err)
	}
	pl, err := plantao.Open(filepath.Join(t.TempDir(), "plantao.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{ScenariosDir: filepath.Join("..", "..", "scenarios"), Store: store, Static: fstest.MapFS{}, Plantao: pl}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)

	post := func(path, body string) (int, map[string]any) {
		res, err := http.Post(ts.URL+path, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var m map[string]any
		_ = json.NewDecoder(res.Body).Decode(&m)
		return res.StatusCode, m
	}

	if code, _ := post("/api/plantao", `{"duracaoMin":60,"ritmo":"normal","trilhas":["inexistente"]}`); code != http.StatusBadRequest {
		t.Errorf("trilha sem cenários: esperava 400, veio %d", code)
	}
	if code, m := post("/api/plantao", `{"duracaoMin":60,"ritmo":"normal","trilhas":["iniciante"]}`); code != http.StatusOK {
		t.Fatalf("assumir: %d %v", code, m)
	}
	if code, _ := post("/api/plantao", `{"duracaoMin":60,"ritmo":"normal","trilhas":["iniciante"]}`); code != http.StatusConflict {
		t.Errorf("segundo turno: esperava 409, veio %d", code)
	}

	res, err := http.Get(ts.URL + "/api/plantao")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if strings.Contains(string(raw), "proximo") {
		t.Error("a API não pode revelar quando o próximo incidente cai")
	}
	var g struct {
		Turno *struct {
			Ritmo   string
			Trilhas []string
		}
		Incidente any
	}
	_ = json.Unmarshal(raw, &g)
	if g.Turno == nil || g.Turno.Ritmo != "normal" || g.Incidente != nil {
		t.Errorf("estado do plantão errado: %s", raw)
	}

	if code, _ := post("/api/plantao/reconhecer", ``); code != http.StatusConflict {
		t.Errorf("reconhecer sem incidente: esperava 409, veio %d", code)
	}
	if code, m := post("/api/plantao/encerrar", ``); code != http.StatusOK {
		t.Errorf("encerrar: %d %v", code, m)
	}
	if code, _ := post("/api/plantao/encerrar", ``); code != http.StatusConflict {
		t.Errorf("encerrar de novo: esperava 409, veio %d", code)
	}
}
