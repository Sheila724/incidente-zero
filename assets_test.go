package incidentezero

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Sheila724/incidente-zero/internal/scenario"
)

func TestCenariosEmbutidos(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "scenarios")
	if err := ExtractScenarios(dst); err != nil {
		t.Fatal(err)
	}
	all, err := scenario.LoadAll(dst)
	if err != nil {
		t.Fatal(err)
	}
	local, _ := scenario.LoadAll("scenarios")
	if len(all) == 0 || len(all) != len(local) {
		t.Fatalf("embutidos: %d cenários, repositório: %d", len(all), len(local))
	}
	st, err := os.Stat(filepath.Join(dst, "iniciante", "001-disco-cheio", "scripts", "check.sh"))
	if err != nil || st.Mode().Perm()&0o100 == 0 {
		t.Error("os scripts extraídos deveriam ser executáveis")
	}
	if len(LabBaseDockerfile) == 0 {
		t.Error("Dockerfile da imagem base não foi embutido")
	}
}
