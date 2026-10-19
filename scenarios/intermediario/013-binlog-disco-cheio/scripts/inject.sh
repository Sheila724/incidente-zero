#!/usr/bin/env bash
# Cenário 013 — injeta a falha:
# a réplica de teste foi desligada, mas o binlog continuou ligado e sem expiração.
# Uma importação noturna de eventos gerou gigabytes de binlog... até a partição encher.
set -euo pipefail
umask 022

livre_mb() { df -Pm /var/lib/mysql | awk 'NR==2 {print $4}'; }

# Lotes de ~1 MB: grava e limpa a tabela de importação. Os dados somem,
# mas cada linha fica registrada no binlog. Passa um pouco do espaço livre.
lotes=$(( $(livre_mb) + 3 ))
{
  echo "USE loja;"
  for i in $(seq 1 "$lotes"); do
    echo "INSERT INTO eventos_importados (lote, payload) VALUES ($i, REPEAT('evento;', 150000));"
    echo "TRUNCATE TABLE eventos_importados;"
  done
} > /root/importacao-eventos.sql

# A importação roda em segundo plano, como o job noturno, e trava quando o disco enche
setsid nohup mariadb < /root/importacao-eventos.sql > /root/importacao-eventos.log 2>&1 &

for _ in $(seq 1 180); do
  [ "$(livre_mb)" -le 1 ] && break
  sleep 1
done
sleep 3

# Autoverificação: a partição do banco cheia e o checkout falhando
if [ "$(livre_mb)" -gt 1 ]; then
  echo "inject: a partição /var/lib/mysql não encheu ($(livre_mb) MB livres)" >&2
  exit 1
fi
code=$(curl -s -o /dev/null -m 15 -w '%{http_code}' http://localhost/checkout || true)
if [ "$code" != "503" ]; then
  echo "inject: falha não reproduzida (checkout HTTP $code)" >&2
  exit 1
fi
