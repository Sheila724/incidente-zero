# Conceitos para este cenário

## Conexões são um recurso finito

Cada cliente que fala com o MariaDB (ou MySQL) abre uma **conexão**. O servidor aceita no máximo `max_connections` ao mesmo tempo. Quando o limite é atingido, todo cliente novo recebe:

```
ERROR 1040 (HY000): Too many connections
```

O banco não está "fora do ar": está cheio. Quem já estava conectado continua funcionando, e quem chega depois é recusado.

```sql
SHOW VARIABLES LIKE 'max_connections';
SHOW STATUS LIKE 'Threads_connected';
SHOW STATUS LIKE 'Max_used_connections';
```

> O MariaDB guarda **uma conexão extra** para quem tem privilégio de administrador. É por isso que `sudo mariadb` ainda entra quando a aplicação já não consegue.

## Quem está conectado

```sql
SHOW PROCESSLIST;
```

| Coluna | Significado |
|---|---|
| `Id` | Identificador da conexão (usado no `KILL`) |
| `User` / `Host` | Quem e de onde |
| `Command` | `Query` = executando; `Sleep` = conectada, sem fazer nada |
| `Time` | Há quantos segundos está nesse estado |

Para enxergar o padrão, agrupe:

```sql
SELECT user, command, COUNT(*) AS conexoes
  FROM information_schema.processlist
 GROUP BY user, command;
```

Muitas conexões em `Sleep` do mesmo usuário é o retrato clássico de um **vazamento de conexões**: o programa abre e não fecha.

## Encerrando conexões

```sql
KILL 123;          -- encerra a conexão 123
```

Matar conexões alivia na hora, mas se o programa que vaza continuar rodando, o banco enche de novo. Reiniciar o banco é a mesma coisa, só que mais brusca: derruba **todo mundo**, inclusive a loja.

## Limites que protegem

| Configuração | Efeito |
|---|---|
| `max_connections` | Limite global do servidor |
| `MAX_USER_CONNECTIONS` (por usuário) | Um usuário não consegue ocupar o banco inteiro |
| `wait_timeout` | Conexões em `Sleep` por mais tempo que isso são fechadas pelo servidor |

```sql
ALTER USER 'relatorio'@'127.0.0.1' WITH MAX_USER_CONNECTIONS 5;
```

Separar usuários por aplicação (a loja usa `loja`, o BI usa `relatorio`) é o que permite enxergar e limitar cada um.

## Abrir e fechar conexões no código

Em Python, o jeito seguro é deixar o `with` fechar a conexão, mesmo se der erro:

```python
with pymysql.connect(...) as conn:
    ...
# aqui a conexão já está fechada
```

Uma conexão aberta durante uma espera longa (rede, FTP, `sleep`) é uma conexão roubada de quem precisa dela.
