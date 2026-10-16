#!/usr/bin/env bash
# Cenário 001 — injeta a falha:
# alguém ligou DEBUG "temporariamente" e esqueceu; o debug.log encheu /var/log.
set -euo pipefail

sed -i 's/^LOG_LEVEL=.*/LOG_LEVEL=DEBUG/' /etc/app/app.conf
sed -i '1i # TEMP: DEBUG ligado para investigar lentidão no checkout. Voltar para INFO depois!' /etc/app/app.conf
systemctl restart app

# O access.log atual tem sobra na última página do tmpfs (4 KB) e continuaria
# aceitando gravações pequenas com o disco cheio. Removendo, a próxima gravação
# precisa de espaço novo e falha como num disco cheio de verdade.
rm -f /var/log/app/access.log

# Simula horas de log acumulado até a partição encher
LINE="$(date -u +%Y-%m-%dT%H:%M:%S) DEBUG cache_dump=$(printf 'x%.0s' {1..900})"
yes "$LINE" > /var/log/app/debug.log 2>/dev/null || true
chown app:app /var/log/app/debug.log
sync

# Autoverificação: a falha precisa estar visível para o aluno
code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/ || true)
if [ "$code" != "503" ]; then
  echo "inject: falha não reproduzida (HTTP $code)" >&2
  exit 1
fi
