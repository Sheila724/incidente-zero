#!/usr/bin/env bash
# Cenário 006 — solução de referência (usada pelo `iz selftest`).
# Nome de arquivo que o cron aceita (sem ponto) e % escapado.
set -euo pipefail

mv /etc/cron.d/backup-loja.cron /etc/cron.d/backup-loja
sed -i 's/\([^\\]\)%/\1\\%/g' /etc/cron.d/backup-loja
