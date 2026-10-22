package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// cmdSelftest testa cada cenário de ponta a ponta:
// sobe o lab com a falha, confere que o check falha, aplica a solução de
// referência (scripts.solve) e confere que o check passa.
func cmdSelftest(args []string) error {
	fs := flag.NewFlagSet("selftest", flag.ContinueOnError)
	espera := fs.Duration("timeout", 4*time.Minute, "tempo máximo para o check passar depois da solução")
	manter := fs.Bool("keep", false, "não destrói o lab no fim (para investigar uma falha)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var alvos []*scenario.Scenario
	if fs.NArg() == 0 {
		all, err := scenario.LoadAll(scenariosDir())
		if err != nil {
			return err
		}
		alvos = all
	} else {
		for _, id := range fs.Args() {
			s, err := scenario.Find(scenariosDir(), id)
			if err != nil {
				return err
			}
			alvos = append(alvos, s)
		}
	}

	type resultado struct {
		id, status string
		dur        time.Duration
	}
	var res []resultado
	falhas := 0
	for _, s := range alvos {
		t0 := time.Now()
		fmt.Printf("\n━━ %s (%s)\n", s.ID, lab.Runtime())
		err := selftestUm(s, *espera)
		status := "✅ ok"
		if err != nil {
			status = "❌ " + err.Error()
			falhas++
			anotarFalha(s, err)
		}
		if !*manter {
			_ = lab.Stop(s)
		}
		res = append(res, resultado{s.ID, status, time.Since(t0).Round(time.Second)})
		fmt.Println(status)
	}

	fmt.Println("\nResumo")
	for _, r := range res {
		fmt.Printf("  %-32s %6s  %s\n", r.id, r.dur, r.status)
	}
	if falhas > 0 {
		return fmt.Errorf("%d de %d cenários falharam", falhas, len(res))
	}
	return nil
}

func selftestUm(s *scenario.Scenario, espera time.Duration) error {
	if s.Scripts.Solve == "" {
		return fmt.Errorf("sem scripts.solve no scenario.yaml")
	}
	passo := func(nome string) { fmt.Printf("  ▶ %s\n", nome) }

	passo("subindo o lab")
	if err := lab.Start(s); err != nil {
		return fmt.Errorf("start: %v", err)
	}
	for _, st := range []struct{ nome, script string }{
		{"setup", s.Scripts.Setup}, {"inject", s.Scripts.Inject},
	} {
		passo(st.nome)
		out, code, err := lab.RunScript(s, st.script)
		ultimaSaida = out
		if err != nil || code != 0 {
			fmt.Print(indentar(out))
			return fmt.Errorf("%s falhou (código %d)", st.nome, code)
		}
	}

	passo("check antes da solução (precisa falhar)")
	out, code, err := lab.RunScript(s, s.Scripts.Check)
	ultimaSaida = out
	if err != nil {
		return fmt.Errorf("check: %v", err)
	}
	if code == 0 {
		fmt.Print(indentar(out))
		return fmt.Errorf("o check passou antes da correção: a falha não foi detectada")
	}

	passo("aplicando a solução de referência")
	out, code, err = lab.RunScript(s, s.Scripts.Solve)
	ultimaSaida = out
	if err != nil || code != 0 {
		fmt.Print(indentar(out))
		return fmt.Errorf("solve falhou (código %d)", code)
	}

	passo("check depois da solução (precisa passar)")
	limite := time.Now().Add(espera)
	for {
		out, code, err = lab.RunScript(s, s.Scripts.Check)
		ultimaSaida = out
		if err == nil && code == 0 {
			return nil
		}
		if time.Now().After(limite) {
			fmt.Print(indentar(out))
			return fmt.Errorf("o check não passou em %s depois da solução", espera)
		}
		time.Sleep(10 * time.Second)
	}
}

// ultimaSaida guarda a última saída de script, para o diagnóstico de uma falha.
var ultimaSaida string

// anotarFalha, no GitHub Actions, publica a falha como anotação (visível no PR
// e pela API), com a saída do check e o estado do lab.
func anotarFalha(s *scenario.Scenario, err error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return
	}
	// Diagnóstico sem o ruído do kernel do host (o journal do lab também recebe essas linhas)
	diag, _, _ := lab.Exec(s, "systemctl --failed --no-legend --plain 2>&1 | head -10\n"+
		"echo '--- journal'; journalctl -n 300 --no-pager -o short 2>&1 | "+
		"grep -vE ' kernel: |fstab-generator|Reloading|EXTRA_OPTS' | tail -40 | cut -c1-180\n", 20*time.Second)
	texto := fmt.Sprintf("%v\n\n%s\n%s", err, strings.TrimSpace(ultimaSaida), diag)
	if len(texto) > 6000 {
		texto = texto[:6000]
	}
	esc := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(texto)
	fmt.Printf("::error title=Selftest %s::%s\n", s.ID, esc)
}

func indentar(s string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	return "    " + strings.ReplaceAll(s, "\n", "\n    ") + "\n"
}
