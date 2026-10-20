package web

import "testing"

func TestCookiePathDoApp(t *testing.T) {
	prefix := "/apps/014-host-sem-dados/zabbix"
	casos := map[string]string{
		"zbx_session=abc; path=/; HttpOnly":   "zbx_session=abc; path=" + prefix + "/; HttpOnly",
		"zbx_session=abc; Path=/":             "zbx_session=abc; Path=" + prefix + "/",
		"tab=1; path=/zabbix/; HttpOnly":      "tab=1; path=/zabbix/; HttpOnly",
		"zbx_session=abc; HttpOnly; SameSite": "zbx_session=abc; HttpOnly; SameSite",
	}
	for entrada, esperado := range casos {
		if got := cookiePath.ReplaceAllString(entrada, "${1}"+prefix+"/${2}"); got != esperado {
			t.Errorf("%q: esperava %q, veio %q", entrada, esperado, got)
		}
	}
}
