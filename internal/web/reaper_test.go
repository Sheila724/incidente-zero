package web

import (
	"reflect"
	"testing"
	"time"
)

func TestIdleLabs(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	idle := 30 * time.Minute
	seen := map[string]time.Time{
		"001-disco-cheio":      now.Add(-31 * time.Minute), // esquecido
		"002-nginx-nao-sobe":   now.Add(-5 * time.Minute),  // em uso
		"003-permissao-negada": now.Add(-2 * time.Hour),    // já foi removido por fora
	}
	running := []string{"001-disco-cheio", "002-nginx-nao-sobe", "004-cpu-100"}

	got := idleLabs(running, seen, now, idle)
	if want := []string{"001-disco-cheio"}; !reflect.DeepEqual(got, want) {
		t.Errorf("expirados: esperava %v, veio %v", want, got)
	}
	if last := seen["004-cpu-100"]; !last.Equal(now) {
		t.Errorf("lab desconhecido deveria começar a contar agora, veio %v", last)
	}
	if _, ok := seen["003-permissao-negada"]; ok {
		t.Error("registro de lab que não roda mais deveria ser limpo")
	}
}
