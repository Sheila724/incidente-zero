// Package usuarios gerencia os logins da demo multiusuário.
//
// Os logins ficam num arquivo de configuração dinâmica do Traefik (o proxy
// com HTTPS e basic auth na frente do `iz ui`). O Traefik relê o arquivo
// sozinho e repassa o nome de quem entrou no cabeçalho X-WebAuth-User.
package usuarios

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Cabecalho é o cabeçalho em que o Traefik informa o usuário autenticado.
const Cabecalho = "X-WebAuth-User"

var reNome = regexp.MustCompile(`^[a-z][a-z0-9-]{1,30}$`)

// NomeValido: minúsculas, números e hífen, começando por letra.
func NomeValido(nome string) bool { return reNome.MatchString(nome) }

const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// APR1 calcula o hash $apr1$ (MD5 do Apache), aceito pelo basic auth do Traefik.
func APR1(senha, sal string) string {
	const magic = "$apr1$"
	if len(sal) > 8 {
		sal = sal[:8]
	}
	pw := []byte(senha)
	s := []byte(sal)

	alt := md5.New()
	alt.Write(pw)
	alt.Write(s)
	alt.Write(pw)
	altSum := alt.Sum(nil)

	ctx := md5.New()
	ctx.Write(pw)
	ctx.Write([]byte(magic))
	ctx.Write(s)
	for i := len(pw); i > 0; i -= 16 {
		n := 16
		if i < 16 {
			n = i
		}
		ctx.Write(altSum[:n])
	}
	for i := len(pw); i > 0; i >>= 1 {
		if i&1 != 0 {
			ctx.Write([]byte{0})
		} else if len(pw) > 0 {
			ctx.Write(pw[:1])
		}
	}
	final := ctx.Sum(nil)

	for i := 0; i < 1000; i++ {
		c := md5.New()
		if i&1 != 0 {
			c.Write(pw)
		} else {
			c.Write(final)
		}
		if i%3 != 0 {
			c.Write(s)
		}
		if i%7 != 0 {
			c.Write(pw)
		}
		if i&1 != 0 {
			c.Write(final)
		} else {
			c.Write(pw)
		}
		final = c.Sum(nil)
	}

	var b strings.Builder
	b.WriteString(magic + sal + "$")
	to64 := func(v uint32, n int) {
		for ; n > 0; n-- {
			b.WriteByte(itoa64[v&0x3f])
			v >>= 6
		}
	}
	for _, g := range [][3]int{{0, 6, 12}, {1, 7, 13}, {2, 8, 14}, {3, 9, 15}, {4, 10, 5}} {
		to64(uint32(final[g[0]])<<16|uint32(final[g[1]])<<8|uint32(final[g[2]]), 4)
	}
	to64(uint32(final[11]), 2)
	return b.String()
}

func aleatorio(alfabeto string, n int) string {
	out := make([]byte, n)
	for i := range out {
		k, err := rand.Int(rand.Reader, big.NewInt(int64(len(alfabeto))))
		if err != nil {
			panic(err)
		}
		out[i] = alfabeto[k.Int64()]
	}
	return string(out)
}

// GerarSenha cria uma senha aleatória fácil de digitar (sem 0/O, 1/l).
func GerarSenha() string {
	const a = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	return aleatorio(a, 4) + "-" + aleatorio(a, 4) + "-" + aleatorio(a, 4)
}

// Hash gera o hash de uma senha com sal aleatório.
func Hash(senha string) string { return APR1(senha, aleatorio(itoa64, 8)) }

// Arquivo é a lista de logins (nome → hash).
type Arquivo struct {
	Caminho string
	Logins  map[string]string
}

var reLinha = regexp.MustCompile(`^\s*-\s*'([a-z][a-z0-9-]{1,30}):(\$apr1\$[^']+)'\s*$`)

// Abrir lê o arquivo (vazio se ainda não existir).
func Abrir(caminho string) (*Arquivo, error) {
	a := &Arquivo{Caminho: caminho, Logins: map[string]string{}}
	f, err := os.Open(caminho)
	if errors.Is(err, os.ErrNotExist) {
		return a, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := reLinha.FindStringSubmatch(sc.Text()); m != nil {
			a.Logins[m[1]] = m[2]
		}
	}
	return a, sc.Err()
}

// Nomes devolve os usuários em ordem alfabética.
func (a *Arquivo) Nomes() []string {
	out := make([]string, 0, len(a.Logins))
	for n := range a.Logins {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Definir cria o usuário ou troca a senha dele.
func (a *Arquivo) Definir(nome, senha string) error {
	if !NomeValido(nome) {
		return fmt.Errorf("nome inválido: %q (use letras minúsculas, números e hífen, começando por letra)", nome)
	}
	if len(senha) < 8 {
		return errors.New("a senha precisa ter pelo menos 8 caracteres")
	}
	a.Logins[nome] = Hash(senha)
	return nil
}

// Remover apaga o usuário.
func (a *Arquivo) Remover(nome string) error {
	if _, ok := a.Logins[nome]; !ok {
		return fmt.Errorf("usuário %q não existe", nome)
	}
	delete(a.Logins, nome)
	return nil
}

// Conteudo é o YAML de configuração dinâmica do Traefik com o middleware iz-auth.
func (a *Arquivo) Conteudo() string {
	var b strings.Builder
	b.WriteString("# Logins do Incidente Zero, gerados pelo `iz usuarios`. Não edite à mão.\n")
	b.WriteString("# O Traefik relê este arquivo sozinho e informa o usuário ao iz em " + Cabecalho + ".\n")
	b.WriteString("http:\n  middlewares:\n    iz-auth:\n      basicAuth:\n")
	b.WriteString("        realm: Incidente Zero\n")
	b.WriteString("        headerField: " + Cabecalho + "\n")
	b.WriteString("        removeHeader: true\n")
	b.WriteString("        users:\n")
	nomes := a.Nomes()
	if len(nomes) == 0 {
		// Traefik exige pelo menos um usuário: um login impossível mantém tudo fechado
		b.WriteString("          - 'ninguem:$apr1$x$bloqueado'\n")
	}
	for _, n := range nomes {
		b.WriteString("          - '" + n + ":" + a.Logins[n] + "'\n")
	}
	return b.String()
}

// Salvar grava o arquivo (atomicamente, na mesma pasta).
func (a *Arquivo) Salvar() error {
	if err := os.MkdirAll(filepath.Dir(a.Caminho), 0o755); err != nil {
		return err
	}
	tmp := a.Caminho + ".tmp"
	if err := os.WriteFile(tmp, []byte(a.Conteudo()), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.Caminho)
}
