# Para estudar mais

## MariaDB

- [Binary log: visão geral](https://mariadb.com/kb/en/overview-of-the-binary-log/)
- [PURGE BINARY LOGS](https://mariadb.com/kb/en/purge-binary-logs/)
- [Usando e mantendo o binlog](https://mariadb.com/kb/en/using-and-maintaining-the-binary-log/): expiração e tamanho
- [binlog_expire_logs_seconds](https://mariadb.com/kb/en/replication-and-binary-log-system-variables/#binlog_expire_logs_seconds): expiração automática

## Monitoramento

- [Template MySQL/MariaDB do Zabbix](https://www.zabbix.com/integrations/mysql) e o item `vfs.fs.size` para a partição do banco

## Para praticar

- Ligue o binlog numa VM de teste, gere alterações e acompanhe `SHOW BINARY LOGS`.
- Restaure um backup e reaplique o binlog até um horário com `mariadb-binlog --stop-datetime`.
- Cenários relacionados: **001 — Disco cheio em /var/log** e **012 — Banco recusando conexões**.
