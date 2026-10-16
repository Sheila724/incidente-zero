#!/usr/bin/env bash
# Cenário 004 — injeta a falha:
# o deploy da tarde zerou o intervalo do gerador de relatórios, que entrou em loop.
set -euo pipefail

cat > /etc/report-worker/worker.conf <<'EOF'
# Gerador de relatórios de vendas
# Intervalo entre relatórios, em segundos
# Ajuste do deploy das 15:30 — testar geração contínua
INTERVAL_SECONDS=0
EOF
systemctl restart report-worker
sleep 3

# Autoverificação: o worker precisa estar consumindo CPU de verdade
pid=$(systemctl show -p MainPID --value report-worker)
a=$(awk '{print $14+$15}' "/proc/$pid/stat")
sleep 2
b=$(awk '{print $14+$15}' "/proc/$pid/stat")
if [ $((b - a)) -lt 30 ]; then
  echo "inject: o worker não está consumindo CPU (${a} -> ${b} ticks)" >&2
  exit 1
fi
