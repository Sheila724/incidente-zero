#!/usr/bin/env bash
# Cenário 008 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

ipct=$(df --output=ipcent /var/lib/app 2>/dev/null | tail -1 | tr -dc 0-9)
if [ -n "$ipct" ] && [ "$ipct" -lt 80 ]; then
  pass "Inodes de /var/lib/app em ${ipct}%"
else
  fail "Inodes de /var/lib/app ainda em ${ipct:-?}%"
fi

if [ -d /var/lib/app/sessions ] && [ "$(stat -c %U /var/lib/app/sessions)" = "app" ]; then
  pass "Pasta de sessões presente e do usuário app"
else
  fail "A pasta /var/lib/app/sessions sumiu ou mudou de dono"
fi

systemctl is-active --quiet app && pass "Serviço app ativo" || fail "Serviço app não está ativo"

code=$(curl -s -o /dev/null -m 5 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-000}"

if systemctl is-enabled --quiet limpa-sessoes.timer && systemctl is-active --quiet limpa-sessoes.timer; then
  pass "Timer limpa-sessoes ativo e habilitado no boot"
else
  fail "O timer limpa-sessoes continua desativado: as sessões vão acumular de novo"
fi

old=$(find /var/lib/app/sessions -type f -mmin +60 2>/dev/null | wc -l)
if [ "$old" -lt 50 ]; then
  pass "Sessões antigas removidas ($old restantes com mais de 60 min)"
else
  fail "Ainda há $old sessões com mais de 60 minutos"
fi

# Bônus: os clientes que estavam comprando continuam logados
list=/var/lib/incidente-zero/sessoes-recentes.txt
if [ -f "$list" ]; then
  kept=0
  while read -r f; do [ -n "$f" ] && [ -e "$f" ] && kept=$((kept + 1)); done < "$list"
  if [ "$kept" -ge 35 ]; then
    pass "Bônus: $kept de 40 sessões recentes preservadas (ninguém foi deslogado)"
  fi
fi

exit $status
