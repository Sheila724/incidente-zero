#!/usr/bin/env bash
# Cenário 006 — injeta a falha:
# a "padronização" renomeou o agendamento com .cron e pôs um % sem escape na linha.
set -euo pipefail

# Histórico: backups até três dias atrás
rm -f /var/backups/loja/loja-*.tar.gz
for d in 3 4 5 6; do
  stamp=$(date -d "-$d days 02:00" +%Y%m%d-%H%M)
  f="/var/backups/loja/loja-$stamp.tar.gz"
  tar -czf "$f" -C /srv/loja data
  touch -d "-$d days 02:00" "$f"
done

rm -f /etc/cron.d/backup-loja
cat > /etc/cron.d/backup-loja.cron <<'EOF'
# Backup dos dados da loja
# Padronizado em segunda-feira: arquivos de agendamento com extensão .cron
# Pedido do time de banco: a cada 2 minutos até validar a rotina (produção: 0 2 * * *)
*/2 * * * * root /usr/local/bin/backup-loja.sh > /var/log/backup-loja-$(date +%F).log 2>&1
EOF
chmod 0644 /etc/cron.d/backup-loja.cron
systemctl restart cron

# Autoverificação
[ -f /etc/cron.d/backup-loja.cron ]
if [ -n "$(find /var/backups/loja -name 'loja-*.tar.gz' -mmin -60 | head -1)" ]; then
  echo "inject: há backups recentes, a falha não foi reproduzida" >&2
  exit 1
fi
