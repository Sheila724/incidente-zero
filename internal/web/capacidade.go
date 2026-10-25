package web

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Sheila724/incidente-zero/internal/lab"
	"github.com/Sheila724/incidente-zero/internal/scenario"
)

// ErrServidorCheio: subir mais um lab passaria do limite de memória do servidor.
var ErrServidorCheio = errors.New("servidor cheio: outros labs estão usando a memória disponível. Tente de novo em alguns minutos")

// ParseMemoria converte valores como "256m", "1792m", "2g" e "512k" em bytes.
func ParseMemoria(v string) (int64, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || v == "0" {
		return 0, nil
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(v, "g"):
		mult, v = 1<<30, strings.TrimSuffix(v, "g")
	case strings.HasSuffix(v, "m"):
		mult, v = 1<<20, strings.TrimSuffix(v, "m")
	case strings.HasSuffix(v, "k"):
		mult, v = 1<<10, strings.TrimSuffix(v, "k")
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n < 0 {
		return 0, errors.New("memória inválida: " + v + " (use, por exemplo, 6g ou 512m)")
	}
	return int64(n * float64(mult)), nil
}

// memoriaDoLab é o limite de memória do lab (512m se o cenário não definir).
func memoriaDoLab(sc *scenario.Scenario) int64 {
	if b, err := ParseMemoria(sc.Resources.Memory); err == nil && b > 0 {
		return b
	}
	return 512 << 20
}

// cabe informa se o lab cabe no limite de memória dos labs, considerando que
// os outros labs do mesmo usuário serão encerrados antes de ele subir.
func (s *Server) cabe(novo *scenario.Scenario) bool {
	if s.LabsMemoria <= 0 {
		return true
	}
	rodando, err := lab.RunningAll()
	if err != nil {
		return true
	}
	total := memoriaDoLab(novo)
	for _, l := range rodando {
		if l.Usuario == novo.Usuario {
			continue
		}
		if sc, err := scenario.Find(s.ScenariosDir, l.ID); err == nil {
			total += memoriaDoLab(sc)
		}
	}
	return total <= s.LabsMemoria
}
