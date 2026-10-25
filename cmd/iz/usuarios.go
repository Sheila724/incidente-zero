package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Sheila724/incidente-zero/internal/usuarios"
)

const usageUsuarios = `Logins da demo multiusuário (Traefik com basic auth na frente do iz ui).

Uso:
  iz usuarios list                 lista os usuários
  iz usuarios add <nome> [senha]   cria o usuário (sem senha, gera uma)
  iz usuarios senha <nome> [senha] troca a senha (sem senha, gera uma)
  iz usuarios rm <nome>            remove o usuário

Opções:
  --arquivo  arquivo de logins do Traefik
             (padrão: $IZ_USUARIOS ou deploy/traefik/conf/usuarios.yml)

O Traefik relê o arquivo sozinho: a mudança vale em poucos segundos.
Nomes: letras minúsculas, números e hífen, começando por letra.
`

func cmdUsuarios(args []string) error {
	fs := flag.NewFlagSet("usuarios", flag.ContinueOnError)
	padrao := os.Getenv("IZ_USUARIOS")
	if padrao == "" {
		padrao = "deploy/traefik/conf/usuarios.yml"
	}
	arquivo := fs.String("arquivo", padrao, "arquivo de logins do Traefik")
	fs.Usage = func() { fmt.Print(usageUsuarios) }
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 || rest[0] == "help" {
		fmt.Print(usageUsuarios)
		return nil
	}
	a, err := usuarios.Abrir(*arquivo)
	if err != nil {
		return fmt.Errorf("lendo %s: %w", *arquivo, err)
	}

	switch rest[0] {
	case "list", "ls":
		nomes := a.Nomes()
		if len(nomes) == 0 {
			fmt.Printf("Nenhum usuário em %s. Crie com: iz usuarios add <nome>\n", *arquivo)
			return nil
		}
		for _, n := range nomes {
			fmt.Println(n)
		}
		return nil

	case "add", "senha":
		if len(rest) < 2 {
			return fmt.Errorf("informe o nome: iz usuarios %s <nome> [senha]", rest[0])
		}
		nome := strings.ToLower(rest[1])
		_, existe := a.Logins[nome]
		if rest[0] == "add" && existe {
			return fmt.Errorf("o usuário %q já existe; para trocar a senha: iz usuarios senha %s", nome, nome)
		}
		if rest[0] == "senha" && !existe {
			return fmt.Errorf("o usuário %q não existe; para criar: iz usuarios add %s", nome, nome)
		}
		senha, gerada := "", false
		if len(rest) > 2 {
			senha = rest[2]
		} else {
			senha, gerada = usuarios.GerarSenha(), true
		}
		if err := a.Definir(nome, senha); err != nil {
			return err
		}
		if err := a.Salvar(); err != nil {
			return err
		}
		if rest[0] == "add" {
			fmt.Printf("✅ Usuário %s criado em %s\n", nome, *arquivo)
		} else {
			fmt.Printf("✅ Senha de %s trocada\n", nome)
		}
		if gerada {
			fmt.Printf("   Senha: %s   (anote: ela não fica guardada em texto)\n", senha)
		}
		avisarTraefik()
		return nil

	case "rm", "remove":
		if len(rest) < 2 {
			return fmt.Errorf("informe o nome: iz usuarios rm <nome>")
		}
		if err := a.Remover(strings.ToLower(rest[1])); err != nil {
			return err
		}
		if err := a.Salvar(); err != nil {
			return err
		}
		fmt.Printf("✅ Usuário %s removido. Os labs dele são encerrados pela limpeza automática.\n", rest[1])
		avisarTraefik()
		return nil
	}
	return fmt.Errorf("comando desconhecido: %s (veja iz usuarios help)", rest[0])
}

// avisarTraefik lembra de conferir a configuração do proxy quando ele não está rodando.
func avisarTraefik() {
	out, err := exec.Command("docker", "ps", "--filter", "name=iz-traefik", "--format", "{{.Names}}").Output()
	if err == nil && strings.TrimSpace(string(out)) == "" {
		fmt.Println("   Atenção: o container iz-traefik não está rodando. Veja deploy/traefik/ no README.")
	}
}
