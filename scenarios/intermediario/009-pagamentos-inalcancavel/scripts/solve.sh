#!/usr/bin/env bash
# Cenário 009 — solução de referência (usada pelo `iz selftest`).
# Remove o IP fixado no /etc/hosts (gravando no próprio arquivo: ele é
# montado pelo Docker) e reinicia a aplicação, que guarda o IP em memória.
set -euo pipefail

grep -v -e 'api\.pagamentos\.interno' -e 'INC-2291' /etc/hosts > /tmp/hosts.novo
cat /tmp/hosts.novo > /etc/hosts
rm -f /tmp/hosts.novo
systemctl restart app
