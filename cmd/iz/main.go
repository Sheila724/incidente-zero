// iz — CLI do Incidente Zero.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	incidentezero "github.com/Sheila724/incidente-zero"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/progress"
	"github.com/Sheila724/incidente-zero/internal/scenario"
	"github.com/Sheila724/incidente-zero/internal/web"
	webui "github.com/Sheila724/incidente-zero/web"
)

// version é preenchida no build de release: -ldflags "-X main.version=v0.1.0"
var version = "dev"

const usage = `Incidente Zero — aprenda a resolver incidentes antes do seu primeiro plantão.

Uso:
  iz ui [--addr host:porta] abre a interface web (padrão 127.0.0.1:8080)
  iz list                  lista os cenários disponíveis
  iz pull                  baixa a imagem base pronta (recomendado)
  iz build                 constrói a imagem base localmente
  iz doctor                confere Docker, runtime e imagem
  iz start <cenário>       sobe o lab, injeta o incidente e abre o terminal
  iz shell [cenário]       reabre o terminal do lab
  iz check [cenário]       verifica se o incidente foi resolvido
  iz hint  [cenário] [n]   mostra a dica n (1, 2, 3...)
  iz concepts [cenário]    mostra os conceitos para estudar
  iz solution [cenário]    mostra a solução comentada e a causa raiz
  iz stop  [cenário]       destrói o lab
  iz selftest [cenários]   testa os cenários de ponta a ponta (para quem cria cenários)
  iz version               mostra a versão

O cenário pode ser o ID completo ou o prefixo: "001" = "001-disco-cheio".
Se omitido, usa o lab em execução.

Variáveis:
  IZ_HOME     pasta com os cenários (padrão: ./scenarios, ou os embutidos no iz)
  IZ_RUNTIME  sysbox | local (padrão: detecta automaticamente)
`

const (
	baseImage   = "ghcr.io/sheila724/incidente-zero/lab-base:latest"
	zabbixImage = "ghcr.io/sheila724/incidente-zero/lab-zabbix:latest"
)

// labImages lista as imagens usadas pelos cenários, começando pela base.
func labImages() []string {
	imgs := []string{baseImage}
	seen := map[string]bool{baseImage: true}
	all, _ := scenario.LoadAll(scenariosDir())
	for _, s := range all {
		if s.Image != "" && !seen[s.Image] {
			seen[s.Image] = true
			imgs = append(imgs, s.Image)
		}
	}
	return imgs
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func home() string {
	if h := os.Getenv("IZ_HOME"); h != "" {
		return h
	}
	return "."
}

var (
	scenOnce sync.Once
	scenDir  string
)

// scenariosDir escolhe de onde ler os cenários:
//  1. IZ_HOME/scenarios, se IZ_HOME estiver definido;
//  2. ./scenarios, quando o iz roda dentro do repositório;
//  3. os cenários embutidos no binário, extraídos no cache do usuário.
func scenariosDir() string {
	scenOnce.Do(func() {
		if os.Getenv("IZ_HOME") != "" {
			scenDir = filepath.Join(home(), "scenarios")
			return
		}
		if m, _ := filepath.Glob(filepath.Join("scenarios", "*", "*", "scenario.yaml")); len(m) > 0 {
			scenDir = "scenarios"
			return
		}
		cache, err := os.UserCacheDir()
		if err != nil {
			cache = os.TempDir()
		}
		dir := filepath.Join(cache, "incidente-zero", "scenarios-"+version)
		if version == "dev" {
			_ = os.RemoveAll(dir)
		}
		if _, err := os.Stat(dir); err != nil {
			if err := incidentezero.ExtractScenarios(dir); err != nil {
				fmt.Fprintln(os.Stderr, "erro ao preparar os cenários:", err)
			}
		}
		scenDir = dir
	})
	return scenDir
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}
	cmd, rest := args[0], args[1:]

	switch cmd {
	case "ui":
		return cmdUI(rest)
	case "list", "ls":
		return cmdList()
	case "build":
		if dir := filepath.Join(home(), "images", "lab-base"); fileExists(filepath.Join(dir, "Dockerfile")) {
			if err := lab.Build(dir, baseImage); err != nil {
				return err
			}
			// Dentro do repositório, constrói também as imagens derivadas (ex.: lab-zabbix)
			if dir := filepath.Join(home(), "images", "lab-zabbix"); fileExists(filepath.Join(dir, "Dockerfile")) {
				return lab.Build(dir, zabbixImage)
			}
			return nil
		}
		return lab.BuildDockerfile(incidentezero.LabBaseDockerfile, baseImage)
	case "pull":
		for _, img := range labImages() {
			if err := lab.Pull(img); err != nil {
				return err
			}
		}
		return nil
	case "doctor":
		return cmdDoctor()
	case "selftest":
		return cmdSelftest(rest)
	case "start":
		if len(rest) < 1 {
			return fmt.Errorf("informe o cenário: iz start 001")
		}
		return cmdStart(rest[0])
	case "shell":
		s, err := resolveRunning(rest)
		if err != nil {
			return err
		}
		return lab.Shell(s)
	case "check":
		return cmdCheck(rest)
	case "hint":
		return cmdHint(rest)
	case "concepts":
		return cmdShow(rest, func(s *scenario.Scenario) []string { return []string{s.Content.Concepts} })
	case "solution":
		return cmdShow(rest, func(s *scenario.Scenario) []string {
			return []string{s.Content.Solution, s.Content.RootCause, s.Content.References}
		})
	case "stop":
		s, err := resolve(rest)
		if err != nil {
			return err
		}
		if err := lab.Stop(s); err != nil {
			return err
		}
		fmt.Printf("Lab %s destruído.\n", s.ID)
		return nil
	case "version", "--version", "-v":
		fmt.Println("iz", version)
		return nil
	case "help", "--help", "-h":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("comando desconhecido: %s (veja `iz help`)", cmd)
	}
}

// resolve usa o cenário informado ou, se omitido, o único lab em execução.
func resolve(rest []string) (*scenario.Scenario, error) {
	if len(rest) > 0 {
		return scenario.Find(scenariosDir(), rest[0])
	}
	ids, err := lab.Running()
	if err != nil {
		return nil, err
	}
	switch len(ids) {
	case 0:
		return nil, fmt.Errorf("nenhum lab em execução; use `iz start <cenário>`")
	case 1:
		return scenario.Find(scenariosDir(), ids[0])
	default:
		return nil, fmt.Errorf("há mais de um lab rodando (%s); informe qual", strings.Join(ids, ", "))
	}
}

// resolveRunning é como resolve, mas exige que o lab esteja de pé.
func resolveRunning(rest []string) (*scenario.Scenario, error) {
	s, err := resolve(rest)
	if err != nil {
		return nil, err
	}
	if !lab.IsRunning(s) {
		return nil, fmt.Errorf("o lab %s não está rodando; use `iz start %s`", s.ID, s.ID)
	}
	return s, nil
}

func cmdList() error {
	all, err := scenario.LoadAll(scenariosDir())
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("nenhum cenário em %s (defina IZ_HOME?)", scenariosDir())
	}
	running, _ := lab.Running()
	isRunning := map[string]bool{}
	for _, id := range running {
		isRunning[id] = true
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNÍVEL\tTRILHA\tTEMPO\tTÍTULO\t")
	for _, s := range all {
		mark := ""
		if isRunning[s.ID] {
			mark = "● rodando"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t~%d min\t%s\t%s\n",
			s.ID, s.Level, s.Track, s.EstimatedMinutes, s.Title, mark)
	}
	return w.Flush()
}

func cmdStart(id string) error {
	s, err := scenario.Find(scenariosDir(), id)
	if err != nil {
		return err
	}

	// Um incidente por vez: encerra outros labs em execução
	if running, _ := lab.Running(); len(running) > 0 {
		for _, rid := range running {
			if rid == s.ID {
				continue
			}
			if other, err := scenario.Find(scenariosDir(), rid); err == nil {
				fmt.Printf("▶ Encerrando o lab %s em execução...\n", rid)
				_ = lab.Stop(other)
			}
		}
	}

	fmt.Printf("▶ Subindo o lab %s (runtime: %s)...\n", s.ID, lab.Runtime())
	if err := lab.Start(s); err != nil {
		return err
	}

	for _, step := range []struct{ name, script string }{
		{"Preparando o ambiente", s.Scripts.Setup},
		{"Injetando o incidente", s.Scripts.Inject},
	} {
		fmt.Printf("▶ %s...\n", step.name)
		out, code, err := lab.RunScript(s, step.script)
		if err != nil || code != 0 {
			fmt.Print(out)
			return fmt.Errorf("%s falhou (código %d): %v", step.script, code, err)
		}
	}

	briefing, err := s.ReadFile(s.Alert.Briefing)
	if err == nil {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(briefing)
		fmt.Println(strings.Repeat("─", 60))
	}
	for _, a := range s.Apps {
		fmt.Printf("ℹ️  Este cenário tem o %s rodando no lab: abra pela interface (`iz ui`), no botão %s", a.Title, a.Title)
		if a.Login != "" {
			fmt.Printf(" (usuário %s, senha %s)", a.Login, a.Password)
		}
		fmt.Println(".")
	}
	fmt.Println("Você está entrando no servidor do incidente. Quando terminar, digite `exit`.")
	fmt.Println("De volta ao seu terminal: `iz check` para validar, `iz hint 1` se travar, `iz stop` para encerrar.")
	fmt.Println()
	return lab.Shell(s)
}

func cmdCheck(rest []string) error {
	s, err := resolveRunning(rest)
	if err != nil {
		return err
	}
	out, code, err := lab.RunScript(s, s.Scripts.Check)
	if err != nil {
		return err
	}
	fmt.Print(out)
	fmt.Println()
	if code == 0 {
		fmt.Println("🎉 Incidente resolvido! Leia a causa raiz com `iz solution` e escreva seu postmortem.")
		return nil
	}
	fmt.Println("Ainda não resolvido. Continue investigando ou use `iz hint 1`.")
	os.Exit(2)
	return nil
}

func cmdHint(rest []string) error {
	// aceita: iz hint | iz hint 2 | iz hint 001 | iz hint 001 2
	n := 1
	var scen []string
	for _, a := range rest {
		if v, err := strconv.Atoi(a); err == nil && len(a) <= 2 {
			n = v
		} else {
			scen = append(scen, a)
		}
	}
	s, err := resolve(scen)
	if err != nil {
		return err
	}
	hints, err := s.Hints()
	if err != nil {
		return err
	}
	if n < 1 || n > len(hints) {
		return fmt.Errorf("este cenário tem dicas de 1 a %d", len(hints))
	}
	h := hints[n-1]
	fmt.Printf("💡 Dica %d de %d (custa %d pontos no modo plantão)\n\n%s\n", n, len(hints), h.Cost, strings.TrimSpace(h.Text))
	return nil
}

func cmdShow(rest []string, files func(*scenario.Scenario) []string) error {
	s, err := resolve(rest)
	if err != nil {
		return err
	}
	for _, f := range files(s) {
		text, err := s.ReadFile(f)
		if err != nil {
			continue
		}
		fmt.Println(text)
		fmt.Println(strings.Repeat("─", 60))
	}
	return nil
}

func cmdUI(args []string) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "endereço de escuta (use 0.0.0.0:8080 atrás de um proxy com autenticação)")
	idle := fs.Duration("lab-idle", 0, "remove labs sem atividade na interface por esse tempo, ex.: 30m (0 desliga)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	store, err := progress.Open(progress.DefaultPath())
	if err != nil {
		return fmt.Errorf("progresso: %w", err)
	}
	srv := &web.Server{
		ScenariosDir: scenariosDir(),
		Store:        store,
		Static:       webui.FS(),
		Version:      version,
	}
	fmt.Printf("Incidente Zero rodando em http://%s  (Ctrl+C para sair)\n", *addr)
	if !strings.HasPrefix(*addr, "127.0.0.1") && !strings.HasPrefix(*addr, "localhost") {
		fmt.Println("Atenção: a interface dá acesso a terminais. Exponha só atrás de um proxy com autenticação.")
	}
	if *idle > 0 {
		fmt.Printf("Labs sem atividade por %s serão removidos automaticamente.\n", *idle)
		go srv.ReapIdle(context.Background(), *idle)
	}
	err = http.ListenAndServe(*addr, srv.Handler())
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func cmdDoctor() error {
	ok := true
	check := func(pass bool, good, bad string) {
		if pass {
			fmt.Println("✅", good)
		} else {
			fmt.Println("❌", bad)
			ok = false
		}
	}
	fmt.Println("iz", version)
	dockerOK := lab.DockerAvailable()
	check(dockerOK, "Docker acessível", "Docker não encontrado ou sem permissão (o usuário está no grupo docker?)")
	if dockerOK {
		rt := lab.Runtime()
		if rt == "sysbox" {
			fmt.Println("✅ Runtime: sysbox (isolamento forte)")
		} else {
			fmt.Println("⚠️  Runtime: local (sem Sysbox; os labs rodam com --privileged)")
		}
		check(lab.ImageExists(baseImage), "Imagem base presente: "+baseImage,
			"Imagem base ausente: rode `iz pull` (ou `iz build`)")
		for _, img := range labImages()[1:] {
			if lab.ImageExists(img) {
				fmt.Println("✅ Imagem presente:", img)
			} else {
				fmt.Println("⚠️  Imagem ausente:", img, "(baixada no primeiro uso, ou rode `iz pull`)")
			}
		}
	}
	all, err := scenario.LoadAll(scenariosDir())
	check(err == nil && len(all) > 0, fmt.Sprintf("%d cenários em %s", len(all), scenariosDir()),
		"Nenhum cenário encontrado")
	if !ok {
		return fmt.Errorf("há itens para corrigir")
	}
	return nil
}
