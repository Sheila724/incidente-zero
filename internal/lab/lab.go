// Package lab cria, controla e remove os containers dos labs via Docker CLI.
package lab

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sheila724/incidente-zero/internal/scenario"
)

const label = "incidente-zero.scenario"

// Runtime decide como o container roda:
//   - "sysbox": runtime sysbox-runc (isolamento forte, só Linux)
//   - "local":  Docker padrão com --privileged (Docker Desktop, WSL2, Mac)
func Runtime() string {
	if r := os.Getenv("IZ_RUNTIME"); r != "" {
		return r
	}
	out, err := exec.Command("docker", "info", "--format", "{{json .Runtimes}}").Output()
	if err == nil && strings.Contains(string(out), "sysbox-runc") {
		return "sysbox"
	}
	return "local"
}

func docker(args ...string) *exec.Cmd {
	return exec.Command("docker", args...)
}

// Start cria o container do cenário e espera o systemd ficar pronto.
func Start(s *scenario.Scenario) error {
	name := s.ContainerName()
	_ = docker("rm", "-f", name).Run()

	args := []string{"run", "-d", "--name", name, "--label", label + "=" + s.ID}
	if s.Hostname != "" {
		args = append(args, "--hostname", s.Hostname)
	}

	switch rt := Runtime(); rt {
	case "sysbox":
		args = append(args, "--runtime=sysbox-runc")
	case "local":
		args = append(args, "--privileged", "--cgroupns=private",
			"--tmpfs", "/run", "--tmpfs", "/run/lock")
	default:
		return fmt.Errorf("IZ_RUNTIME inválido: %q (use sysbox ou local)", rt)
	}

	r := s.Resources
	if r.CPUs > 0 {
		args = append(args, "--cpus", strconv.FormatFloat(r.CPUs, 'f', -1, 64))
	}
	if r.Memory != "" {
		args = append(args, "--memory", r.Memory)
	}
	if r.PIDs > 0 {
		args = append(args, "--pids-limit", strconv.Itoa(r.PIDs))
	}
	for _, m := range r.Mounts {
		if m.Type != "tmpfs" {
			return fmt.Errorf("tipo de mount não suportado: %s", m.Type)
		}
		if m.Inodes > 0 {
			// --mount não aceita nr_inodes; --tmpfs aceita as opções do kernel
			opts := "rw,mode=755,nr_inodes=" + strconv.Itoa(m.Inodes)
			if m.Size != "" {
				opts += ",size=" + m.Size
			}
			args = append(args, "--tmpfs", m.Path+":"+opts)
			continue
		}
		mount := "type=tmpfs,destination=" + m.Path
		if m.Size != "" {
			mount += ",tmpfs-size=" + m.Size
		}
		args = append(args, "--mount", mount)
	}
	args = append(args, s.Image)

	if out, err := docker(args...).CombinedOutput(); err != nil {
		return fmt.Errorf("docker run falhou: %s", strings.TrimSpace(string(out)))
	}
	return waitSystemd(name, 60*time.Second)
}

func waitSystemd(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, _ := docker("exec", name, "systemctl", "is-system-running").Output()
		state := strings.TrimSpace(string(out))
		if state == "running" || state == "degraded" {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("o systemd do lab não ficou pronto em %s", timeout)
}

// RunScript executa um script do cenário como root dentro do lab.
// Retorna a saída combinada e o código de saída.
func RunScript(s *scenario.Scenario, rel string) (string, int, error) {
	f, err := os.Open(s.Path(rel))
	if err != nil {
		return "", -1, err
	}
	defer f.Close()

	var buf bytes.Buffer
	cmd := docker("exec", "-i", s.ContainerName(), "bash", "-s")
	cmd.Stdin = f
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return buf.String(), exitErr.ExitCode(), nil
	}
	if err != nil {
		return buf.String(), -1, err
	}
	return buf.String(), 0, nil
}

// Shell abre um terminal interativo no lab como o usuário do cenário.
func Shell(s *scenario.Scenario) error {
	cmd := docker("exec", "-it", "-u", s.User, "-w", "/home/"+s.User,
		s.ContainerName(), "bash", "-l")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		return nil // sair do shell com código != 0 não é erro da ferramenta
	}
	return err
}

// IsRunning informa se o lab do cenário está em execução.
func IsRunning(s *scenario.Scenario) bool {
	out, err := docker("inspect", "-f", "{{.State.Running}}", s.ContainerName()).Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// Stop remove o container do lab.
func Stop(s *scenario.Scenario) error {
	out, err := docker("rm", "-f", s.ContainerName()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Running devolve os IDs dos cenários com lab em execução.
func Running() ([]string, error) {
	out, err := docker("ps", "--filter", "label="+label,
		"--format", `{{.Label "`+label+`"}}`).Output()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			ids = append(ids, l)
		}
	}
	return ids, nil
}

// Build constrói a imagem base dos labs.
func Build(contextDir, tag string) error {
	cmd := docker("build", "-t", tag, contextDir)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// BuildDockerfile constrói a imagem a partir de um Dockerfile em memória (sem contexto).
func BuildDockerfile(dockerfile []byte, tag string) error {
	cmd := docker("build", "-t", tag, "-")
	cmd.Stdin = bytes.NewReader(dockerfile)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Pull baixa a imagem pronta do registry.
func Pull(tag string) error {
	cmd := docker("pull", tag)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// IP devolve o endereço do container do lab na rede do Docker.
func IP(s *scenario.Scenario) (string, error) {
	out, err := docker("inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}",
		s.ContainerName()).Output()
	if err != nil {
		return "", fmt.Errorf("lab não encontrado")
	}
	for _, ip := range strings.Fields(string(out)) {
		if ip != "" {
			return ip, nil
		}
	}
	return "", fmt.Errorf("o lab não tem endereço de rede")
}

// ImageExists informa se a imagem já está no Docker local.
func ImageExists(tag string) bool {
	return docker("image", "inspect", tag).Run() == nil
}

// DockerAvailable informa se o daemon do Docker responde.
func DockerAvailable() bool {
	return docker("info", "--format", "{{.ServerVersion}}").Run() == nil
}
