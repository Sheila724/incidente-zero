package main

import (
	"flag"
	"fmt"
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
		if err != nil || code != 0 {
			fmt.Print(indentar(out))
			return fmt.Errorf("%s falhou (código %d)", st.nome, code)
		}
	}

	passo("check antes da solução (precisa falhar)")
	out, code, err := lab.RunScript(s, s.Scripts.Check)
	if err != nil {
		return fmt.Errorf("check: %v", err)
	}
	if code == 0 {
		fmt.Print(indentar(out))
		return fmt.Errorf("o check passou antes da correção: a falha não foi detectada")
	}

	passo("aplicando a solução de referência")
	out, code, err = lab.RunScript(s, s.Scripts.Solve)
	if err != nil || code != 0 {
		fmt.Print(indentar(out))
		return fmt.Errorf("solve falhou (código %d)", code)
	}

	passo("check depois da solução (precisa passar)")
	limite := time.Now().Add(espera)
	for {
		out, code, err = lab.RunScript(s, s.Scripts.Check)
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

func indentar(s string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	return "    " + strings.ReplaceAll(s, "\n", "\n    ") + "\n"
}
