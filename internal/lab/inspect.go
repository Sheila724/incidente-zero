package lab

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// Exec roda um trecho de shell como root dentro do lab, com tempo limite.
func Exec(s *scenario.Scenario, script string, timeout time.Duration) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", s.ContainerName(), "bash", "-s")
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if ctx.Err() != nil {
		return buf.String(), -1, fmt.Errorf("tempo esgotado após %s", timeout)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return buf.String(), exitErr.ExitCode(), nil
	}
	if err != nil {
		return buf.String(), -1, err
	}
	return buf.String(), 0, nil
}

// ShellCommand monta o comando do terminal interativo do aluno (usado com PTY).
func ShellCommand(s *scenario.Scenario) *exec.Cmd {
	return exec.Command("docker", "exec", "-it",
		"-u", s.User, "-w", "/home/"+s.User,
		"-e", "TERM=xterm-256color",
		s.ContainerName(), "bash", "-l")
}

// Metrics é o retrato do lab exibido no painel de monitoramento.
type Metrics struct {
	HTTPCode    string  `json:"httpCode"`
	Port80      bool    `json:"port80"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float64 `json:"memPercent"`
	MemUsage    string  `json:"memUsage"`
	DiskPercent int     `json:"diskPercent"`
	DiskPath    string  `json:"diskPath"`
	At          int64   `json:"at"`
}

const probeScript = `
code=$(curl -s -o /dev/null -m 2 -w '%{http_code}' http://localhost/ 2>/dev/null)
echo "http=${code:-000}"
if ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq '(^|:)80$'; then echo "port80=1"; else echo "port80=0"; fi
p=/var/log; mountpoint -q /var/log 2>/dev/null || p=/
echo "diskpath=$p"
echo "disk=$(df --output=pcent "$p" 2>/dev/null | tail -1 | tr -dc 0-9)"
`

// Collect lê HTTP, porta 80 e disco de dentro do lab, e CPU/memória do docker stats.
func Collect(s *scenario.Scenario) (Metrics, error) {
	m := Metrics{At: time.Now().UnixMilli(), HTTPCode: "000"}

	out, _, err := Exec(s, probeScript, 6*time.Second)
	if err != nil {
		return m, err
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok {
			continue
		}
		switch k {
		case "http":
			m.HTTPCode = v
		case "port80":
			m.Port80 = v == "1"
		case "disk":
			m.DiskPercent, _ = strconv.Atoi(v)
		case "diskpath":
			m.DiskPath = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	st, err := exec.CommandContext(ctx, "docker", "stats", "--no-stream",
		"--format", "{{.CPUPerc}};{{.MemPerc}};{{.MemUsage}}", s.ContainerName()).Output()
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(st)), ";", 3)
		if len(parts) == 3 {
			m.CPUPercent = pct(parts[0])
			m.MemPercent = pct(parts[1])
			m.MemUsage = strings.TrimSpace(parts[2])
		}
	}
	return m, nil
}

func pct(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(s), "%"), 64)
	return v
}

// StepStatus executa as verificações do modo guiado e diz quais passaram.
// Passos sem "verify" voltam como nil (o aluno marca por conta própria).
func StepStatus(s *scenario.Scenario) ([]*bool, error) {
	var b strings.Builder
	for i, st := range s.Guided {
		if strings.TrimSpace(st.Verify) == "" {
			continue
		}
		fmt.Fprintf(&b, "if ( %s\n) >/dev/null 2>&1; then echo \"s%d=1\"; else echo \"s%d=0\"; fi\n", st.Verify, i, i)
	}
	res := make([]*bool, len(s.Guided))
	if b.Len() == 0 {
		return res, nil
	}
	out, _, err := Exec(s, b.String(), 20*time.Second)
	if err != nil {
		return res, err
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok || !strings.HasPrefix(k, "s") {
			continue
		}
		i, err := strconv.Atoi(k[1:])
		if err != nil || i < 0 || i >= len(res) {
			continue
		}
		ok2 := v == "1"
		res[i] = &ok2
	}
	return res, nil
}
