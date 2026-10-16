#!/usr/bin/env bash
# Cenário 001 — valida a correção. Exit 0 = resolvido.
# Verifica a causa raiz (nível de log em execução), não só o sintoma.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

uso=$(df --output=pcent /var/log | tail -1 | tr -dc '0-9')
if [ "${uso:-100}" -lt 80 ]; then
  pass "Espaço em /var/log normalizado (${uso}% em uso)"
else
  fail "/var/log ainda está com ${uso}% de uso"
fi

systemctl is-active --quiet app   && pass "Serviço app ativo"   || fail "Serviço app não está ativo"
systemctl is-active --quiet nginx && pass "Serviço nginx ativo" || fail "Serviço nginx não está ativo"

code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Site respondendo HTTP 200" || fail "Site respondendo HTTP ${code:-sem resposta}"

level=$(curl -s http://127.0.0.1:8080/health \
  | python3 -c 'import sys, json; print(json.load(sys.stdin)["log_level"])' 2>/dev/null)
if [ -n "$level" ] && [ "$level" != "DEBUG" ]; then
  pass "Aplicação rodando com LOG_LEVEL=$level"
else
  fail "A aplicação em execução ainda está em DEBUG: o disco vai encher de novo"
fi

if [ -f /etc/logrotate.d/app ]; then
  pass "Bônus: rotação de logs configurada"
fi

exit $status
