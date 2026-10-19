// Package incidentezero embute no binário iz os cenários e o Dockerfile da
// imagem base, para que o iz funcione baixado sozinho, sem clonar o repositório.
package incidentezero

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:scenarios
var scenarios embed.FS

//go:embed images/lab-base/Dockerfile
var LabBaseDockerfile []byte

// ExtractScenarios copia os cenários embutidos para dst (que não pode existir).
// Escreve numa pasta temporária e renomeia no fim, para nunca deixar uma cópia pela metade.
func ExtractScenarios(dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dst), ".scenarios-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	err = fs.WalkDir(scenarios, "scenarios", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("scenarios", p)
		target := filepath.Join(tmp, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := scenarios.ReadFile(p)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if filepath.Ext(p) == ".sh" {
			mode = 0o755
		}
		return os.WriteFile(target, b, mode)
	})
	if err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
