#!/usr/bin/env bash
# Cenário 014 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

sed -i -e 's/^Server=.*/Server=127.0.0.1/' -e 's/^ServerActive=.*/ServerActive=127.0.0.1/' \
  /etc/zabbix/zabbix_agent2.conf
systemctl restart zabbix-agent2
rm -f /var/log/auditoria/export-completo-*.json
