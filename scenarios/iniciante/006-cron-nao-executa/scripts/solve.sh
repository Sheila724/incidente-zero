#!/usr/bin/env bash
# Cenário 006 — solução de referência (usada pelo `iz selftest`).
# Nome de arquivo que o cron aceita (sem ponto) e % escapado.
set -euo pipefail

# O cron só relê o /etc/cron.d quando o horário de modificação da pasta muda,
# com precisão de segundos. Num teste automático, a correção pode cair no mesmo
# segundo em que o cron carregou os arquivos e passar despercebida.
sleep 2

mv /etc/cron.d/backup-loja.cron /etc/cron.d/backup-loja
sed -i 's/\([^\\]\)%/\1\\%/g' /etc/cron.d/backup-loja
