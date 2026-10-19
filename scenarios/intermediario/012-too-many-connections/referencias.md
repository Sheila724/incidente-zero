# Para estudar mais

## MariaDB

- [SHOW PROCESSLIST](https://mariadb.com/kb/en/show-processlist/): conexões e o que cada uma está fazendo
- [KILL](https://mariadb.com/kb/en/kill/): encerrar uma conexão ou consulta
- [max_connections e variáveis do servidor](https://mariadb.com/kb/en/server-system-variables/#max_connections)
- [Limites por usuário (MAX_USER_CONNECTIONS)](https://mariadb.com/kb/en/create-user/#resource-limit-options)

## Monitoramento

- [Template MySQL/MariaDB do Zabbix](https://www.zabbix.com/integrations/mysql): inclui conexões em uso e erros de conexão

## Para praticar

- Abra várias sessões com `mariadb` e acompanhe `SHOW STATUS LIKE 'Threads_connected'`.
- Crie um usuário com `MAX_USER_CONNECTIONS 2` e tente abrir a terceira sessão.
- Próximo cenário relacionado: **013 — Banco travado por disco cheio**.
