package usuarios

import (
	"path/filepath"
	"strings"
	"testing"
)

// Vetores gerados com `openssl passwd -apr1 -salt <sal> <senha>`.
func TestAPR1(t *testing.T) {
	casos := []struct{ senha, sal, quer string }{
		{"minhaSenha!", "r31abcde", "$apr1$r31abcde$OHXNcBQZsoCN6ezW87Ah9/"},
		{"a", "Zx9", "$apr1$Zx9$aKNDKAHah7CCUUsfpaVEK0"},
		{"senha com espaço e ç", "12345678", "$apr1$12345678$G8as8EcVayJGSqIdk6VYE0"},
	}
	for _, c := range casos {
		if got := APR1(c.senha, c.sal); got != c.quer {
			t.Errorf("APR1(%q, %q) = %s, quer %s", c.senha, c.sal, got, c.quer)
		}
	}
}

func TestArquivo(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "conf", "usuarios.yml")
	a, err := Abrir(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.Conteudo(), "ninguem:") {
		t.Error("arquivo sem usuários precisa continuar fechado")
	}
	if err := a.Definir("Ana", "12345678"); err == nil {
		t.Error("aceitou nome com maiúscula")
	}
	if err := a.Definir("ana", "curta"); err == nil {
		t.Error("aceitou senha curta")
	}
	if err := a.Definir("ana", "uma-senha-boa"); err != nil {
		t.Fatal(err)
	}
	if err := a.Definir("bia", GerarSenha()); err != nil {
		t.Fatal(err)
	}
	if err := a.Salvar(); err != nil {
		t.Fatal(err)
	}

	b, err := Abrir(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(b.Nomes(), ",") != "ana,bia" {
		t.Fatalf("usuários relidos: %v", b.Nomes())
	}
	// O hash gravado confere com a senha
	h := b.Logins["ana"]
	sal := strings.Split(h, "$")[2]
	if APR1("uma-senha-boa", sal) != h {
		t.Error("hash gravado não confere com a senha")
	}
	if !strings.Contains(b.Conteudo(), "headerField: "+Cabecalho) {
		t.Error("o Traefik precisa repassar o usuário no cabeçalho")
	}
	if err := b.Remover("bia"); err != nil {
		t.Fatal(err)
	}
	if err := b.Remover("bia"); err == nil {
		t.Error("removeu usuário inexistente")
	}
}

func TestGerarSenha(t *testing.T) {
	s := GerarSenha()
	if len(s) != 14 || strings.Count(s, "-") != 2 {
		t.Errorf("senha gerada fora do formato: %q", s)
	}
	if GerarSenha() == s {
		t.Error("duas senhas iguais")
	}
}
