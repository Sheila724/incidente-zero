#!/usr/bin/env bash
# Cenário 015 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-sem resposta}"

if systemctl is-active --quiet trafego-clientes; then
  pass "Clientes continuam navegando (tráfego ativo)"
else
  contorno "O tráfego de clientes foi desligado: tirar os clientes não é solução"
fi

if grep -q '^APP_VERSION=2.5.0' /etc/app/app.env; then
  pass "Versão 2.5.0 mantida (o cache de sessões continua no ar)"
else
  contorno "A versão 2.5.0 foi desfeita: o objetivo é corrigir a configuração, não fazer rollback"
fi

limite=$(curl -s -m 5 http://127.0.0.1:8080/health | jq -r '.session_cache_max // empty' 2>/dev/null)
if [ -n "$limite" ] && [ "$limite" -gt 0 ] && [ "$limite" -le 10000 ]; then
  pass "A loja em execução limita o cache de sessões (SESSION_CACHE_MAX=$limite)"
else
  fail "A loja em execução não tem limite de sessões em memória (SESSION_CACHE_MAX=${limite:-?})"
fi

pid=$(systemctl show -p MainPID --value app)
rss_mb=$(( $(ps -o rss= -p "$pid" 2>/dev/null || echo 999999) / 1024 ))
if [ "$rss_mb" -lt 250 ]; then
  pass "Memória da loja em ${rss_mb} MB"
else
  fail "A loja continua usando ${rss_mb} MB de memória"
fi

problemas=$(zbx-api problem.get '{"search":{"name":"Loja Exemplo"},"output":["name"]}' 2>/dev/null | jq -r '.[].name')
if [ -z "$problemas" ]; then
  pass "Nenhum incidente da Loja Exemplo aberto no Zabbix"
else
  while IFS= read -r p; do
    [ -n "$p" ] && fail "Incidente ainda aberto no Zabbix: $p"
  done <<< "$problemas"
fi

trig=$(zbx-api trigger.get '{"search":{"description":"usando muita memória"},"output":["triggerid"]}' 2>/dev/null | jq -r '.[0].triggerid // empty')
msg=""
if [ -n "$trig" ]; then
  msg=$(zbx-api event.get "{\"objectids\":\"$trig\",\"source\":0,\"object\":0,\"value\":1,
    \"sortfield\":[\"clock\"],\"sortorder\":\"DESC\",\"limit\":1,
    \"select_acknowledges\":[\"message\"],\"output\":[\"eventid\"]}" 2>/dev/null \
    | jq -r '[.[0].acknowledges[]?.message | select(length > 0)] | first // empty')
fi
if [ -n "$msg" ]; then
  pass "Incidente comentado no Zabbix: \"$msg\""
else
  fail "Registre no Zabbix o que foi feito: em Incidentes, clique em Atualizar e escreva uma mensagem"
fi

exit $status
