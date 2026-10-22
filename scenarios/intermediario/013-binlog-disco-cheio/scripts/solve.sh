#!/usr/bin/env bash
# Cenário 013 — solução de referência (usada pelo `iz selftest`).
# 1. libera espaço no sistema de arquivos (o MariaDB está travado esperando);
# 2. espera o servidor voltar; 3. PURGE; 4. expiração do binlog.
set -euo pipefail

mapfile -t antigos < <(ls /var/lib/mysql/mysql-bin.[0-9]* | sort | head -2)
mv "${antigos[@]}" /root/

for _ in $(seq 1 90); do
  mariadb -e 'SHOW PROCESSLIST' | grep -q 'free space' || break
  sleep 2
done

atual=$(mariadb -N -e 'SHOW MASTER STATUS' | awk '{print $1}')
mariadb -e "PURGE BINARY LOGS TO '$atual'"
rm -f /root/mysql-bin.*

mariadb -e 'SET GLOBAL binlog_expire_logs_seconds = 86400'
sed -i 's/^binlog_expire_logs_seconds = 0/binlog_expire_logs_seconds = 86400/' \
  /etc/mysql/mariadb.conf.d/60-loja.cnf
