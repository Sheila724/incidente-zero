#!/usr/bin/env bash
# Cenário 014 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

if [ "$(zabbix_get -s 127.0.0.1 -k agent.ping 2>/dev/null)" = "1" ]; then
  pass "O agente responde ao Zabbix Server (zabbix_get agent.ping = 1)"
else
  fail "O agente ainda não responde ao Zabbix Server em 127.0.0.1"
fi

server=$(grep -E '^Server=' /etc/zabbix/zabbix_agent2.conf | tail -1 | cut -d= -f2)
if echo "$server" | grep -Eq '(^|,)0\.0\.0\.0(/0)?(,|$)|(^|,)::/0(,|$)'; then
  fail "Server=$server libera o agente para qualquer endereço: use o IP do Zabbix Server"
elif echo "$server" | grep -Eq '(^|,)(127\.0\.0\.1|localhost)(,|$)'; then
  pass "Server= aponta para o Zabbix Server ($server)"
else
  fail "Server= ($server) não inclui o Zabbix Server (127.0.0.1)"
fi

host=$(zbx-api host.get '{"filter":{"host":["loja-web-01"]},"output":["hostid"],"selectInterfaces":["available"]}' 2>/dev/null)
disp=$(echo "$host" | jq -r '.[0].interfaces[0].available // empty')
if [ "$disp" = "1" ]; then
  pass "No Zabbix, a interface do agente de loja-web-01 está disponível"
else
  fail "No Zabbix, o agente de loja-web-01 ainda aparece indisponível"
fi

uso=$(df --output=pcent /var/log | tail -1 | tr -dc '0-9')
if [ "${uso:-100}" -le 90 ]; then
  pass "/var/log com ${uso}% de uso"
else
  fail "/var/log com ${uso}% de uso: veja o que o Zabbix passou a mostrar depois que voltou a receber dados"
fi

hostid=$(echo "$host" | jq -r '.[0].hostid // empty')
problemas=$(zbx-api problem.get "{\"hostids\":\"$hostid\",\"output\":[\"name\"]}" 2>/dev/null | jq -r '.[].name')
if [ -n "$hostid" ] && [ -z "$problemas" ]; then
  pass "Nenhum problema aberto para loja-web-01 no Zabbix"
else
  while IFS= read -r p; do
    [ -n "$p" ] && fail "Problema ainda aberto no Zabbix: $p"
  done <<< "$problemas"
  [ -z "$hostid" ] && fail "Não consegui consultar o Zabbix"
fi

exit $status
