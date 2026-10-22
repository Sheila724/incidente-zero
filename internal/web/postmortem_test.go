package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Sheila724/incidente-zero/internal/progress"
)

func TestPostmortem(t *testing.T) {
	store, err := progress.Open(filepath.Join(t.TempDir(), "p.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{ScenariosDir: filepath.Join("..", "..", "scenarios"), Store: store, Static: fstest.MapFS{}}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	put := func(body string) int {
		req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/scenarios/001/postmortem", strings.NewReader(body))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}

	if code := put(`{"causa":"x"}`); code != http.StatusConflict {
		t.Errorf("antes de resolver: esperava 409, veio %d", code)
	}

	_ = store.Start("001-disco-cheio")
	_ = store.UseHint("001-disco-cheio", 1, 10)
	_, _ = store.Resolve("001-disco-cheio", 100, 10)

	if code := put(`{"causa":"  LOG_LEVEL=DEBUG esquecido  ","acoes":"- logrotate\nalerta de disco em 80%"}`); code != http.StatusOK {
		t.Fatalf("salvar: esperava 200, veio %d", code)
	}

	res, err := http.Get(ts.URL + "/api/scenarios/001/postmortem.md?tz=America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, res.Body)
	md := buf.String()
	for _, esperado := range []string{
		"# Postmortem: Disco cheio em /var/log",
		"LOG_LEVEL=DEBUG esquecido\n",
		"- [ ] logrotate",
		"- [ ] alerta de disco em 80%",
		"Plantão assumido:",
		"Dica 1 revelada",
		"- **Severidade:** Alta",
		"(-03)",
	} {
		if !strings.Contains(md, esperado) {
			t.Errorf("markdown sem %q:\n%s", esperado, md)
		}
	}
	if !strings.Contains(res.Header.Get("Content-Disposition"), "postmortem-001-disco-cheio.md") {
		t.Error("download sem nome de arquivo")
	}
}
