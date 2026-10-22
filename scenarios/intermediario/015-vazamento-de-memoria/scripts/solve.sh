#!/usr/bin/env bash
# Cenário 015 — solução de referência (usada pelo `iz selftest`).
# Limita o cache de sessões, reinicia a loja e registra o incidente no Zabbix.
set -euo pipefail

sed -i 's/^SESSION_CACHE_MAX=0/SESSION_CACHE_MAX=2000/' /etc/app/app.env
systemctl restart app

trig=$(zbx-api trigger.get '{"search":{"description":"usando muita memória"},"output":["triggerid"]}' \
  | jq -r '.[0].triggerid')
evento=$(zbx-api event.get "{\"objectids\":\"$trig\",\"value\":1,\"sortfield\":[\"clock\"],
  \"sortorder\":\"DESC\",\"limit\":1,\"output\":[\"eventid\"]}" | jq -r '.[0].eventid')
zbx-api event.acknowledge "{\"eventids\":\"$evento\",\"action\":6,
  \"message\":\"Cache de sessões sem limite na 2.5.0: SESSION_CACHE_MAX=2000 e loja reiniciada.\"}" >/dev/null
