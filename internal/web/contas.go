package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Sheila724/incidente-zero/internal/plantao"
	"github.com/Sheila724/incidente-zero/internal/progress"
	"github.com/Sheila724/incidente-zero/internal/usuarios"
)

// Modo multiusuário: o proxy na frente do iz (Traefik com basic auth) informa
// quem é a pessoa num cabeçalho (ex.: X-WebAuth-User). Cada usuário tem o seu
// progresso, o seu plantão e os seus labs. Sem o cabeçalho configurado, o iz
// funciona como sempre, para uma pessoa só.

// conta é o espaço de uma pessoa: progresso e plantão.
type conta struct {
	Nome    string // vazio no uso local (uma pessoa só)
	Store   *progress.Store
	Plantao *plantao.Plantao
}

// NomeUsuarioValido informa se o nome serve para usuário: minúsculas, números
// e hífen, começando por letra (vira parte do nome do container e de pastas).
func NomeUsuarioValido(nome string) bool { return usuarios.NomeValido(nome) }

func (s *Server) multiusuario() bool { return s.UsuarioHeader != "" }

type chaveConta struct{}

// contaReq devolve a conta de quem fez o pedido.
func contaReq(r *http.Request) *conta {
	c, _ := r.Context().Value(chaveConta{}).(*conta)
	return c
}

// comConta identifica a pessoa em cada pedido da API e das interfaces dos labs.
func (s *Server) comConta(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/apps/") {
			next.ServeHTTP(w, r)
			return
		}
		nome := ""
		if s.multiusuario() {
			nome = strings.ToLower(strings.TrimSpace(r.Header.Get(s.UsuarioHeader)))
			if nome == "" {
				writeErr(w, http.StatusUnauthorized, "usuário não identificado: o proxy precisa enviar o cabeçalho "+s.UsuarioHeader)
				return
			}
			if !NomeUsuarioValido(nome) {
				writeErr(w, http.StatusForbidden, "nome de usuário inválido: use letras minúsculas, números e hífen")
				return
			}
		}
		c, err := s.abrirConta(nome)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), chaveConta{}, c)))
	})
}

func (s *Server) pastaUsuarios() string { return filepath.Join(s.DataDir, "usuarios") }

// abrirConta devolve (e guarda em memória) a conta de um usuário.
func (s *Server) abrirConta(nome string) (*conta, error) {
	if nome == "" {
		return &conta{Store: s.Store, Plantao: s.Plantao}, nil
	}
	s.contasMu.Lock()
	defer s.contasMu.Unlock()
	if c, ok := s.contasAbertas[nome]; ok {
		return c, nil
	}
	dir := filepath.Join(s.pastaUsuarios(), nome)
	st, err := progress.Open(filepath.Join(dir, "progress.json"))
	if err != nil {
		return nil, fmt.Errorf("progresso de %s: %w", nome, err)
	}
	pl, err := plantao.Open(filepath.Join(dir, "plantao.json"))
	if err != nil {
		return nil, fmt.Errorf("plantão de %s: %w", nome, err)
	}
	c := &conta{Nome: nome, Store: st, Plantao: pl}
	if s.contasAbertas == nil {
		s.contasAbertas = map[string]*conta{}
	}
	s.contasAbertas[nome] = c
	return c, nil
}

// contas lista as contas conhecidas: a local ou, no modo multiusuário, todas
// as que têm pasta (assim um plantão em andamento continua depois de um restart).
func (s *Server) contas() []*conta {
	if !s.multiusuario() {
		return []*conta{{Store: s.Store, Plantao: s.Plantao}}
	}
	if entradas, err := os.ReadDir(s.pastaUsuarios()); err == nil {
		for _, e := range entradas {
			if e.IsDir() && NomeUsuarioValido(e.Name()) {
				if _, err := s.abrirConta(e.Name()); err != nil {
					log.Printf("contas: %v", err)
				}
			}
		}
	}
	s.contasMu.Lock()
	defer s.contasMu.Unlock()
	out := make([]*conta, 0, len(s.contasAbertas))
	for _, c := range s.contasAbertas {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nome < out[j].Nome })
	return out
}

// contaPorNome devolve a conta já aberta de um usuário ("" = local).
func (s *Server) contaPorNome(nome string) *conta {
	c, err := s.abrirConta(nome)
	if err != nil {
		return nil
	}
	return c
}
