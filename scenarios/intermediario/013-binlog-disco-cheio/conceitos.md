# Conceitos para este cenário

## Binary log (binlog)

O binlog é um registro de **todas as alterações** feitas no banco (INSERT, UPDATE, DELETE...), em ordem. Ele serve para:

- **Replicação:** a réplica lê o binlog do servidor principal e repete as mudanças.
- **Recuperação no tempo (point-in-time):** restaurar um backup e reaplicar o binlog até um momento exato.

Ele fica em arquivos numerados (`mysql-bin.000001`, `000002`...), com um índice (`mysql-bin.index`) que lista todos. Um arquivo novo começa quando o atual chega em `max_binlog_size` ou o servidor reinicia.

> Os dados apagados de uma tabela **continuam no binlog**. Uma importação que grava e limpa a mesma tabela milhares de vezes deixa a tabela pequena e o binlog enorme.

## Quanto tempo o binlog fica guardado

| Configuração | Efeito |
|---|---|
| `binlog_expire_logs_seconds = 0` | Nunca apaga nada: cresce até encher o disco |
| `binlog_expire_logs_seconds = 86400` | Apaga sozinho arquivos com mais de 1 dia |
| `expire_logs_days` | O mesmo, em dias (nome antigo) |

```sql
SHOW BINARY LOGS;          -- arquivos e tamanhos
SHOW MASTER STATUS;        -- arquivo e posição atuais
PURGE BINARY LOGS TO 'mysql-bin.000014';     -- apaga os anteriores a esse
PURGE BINARY LOGS BEFORE NOW() - INTERVAL 1 DAY;
```

## Por que não usar `rm`

O `PURGE` apaga os arquivos **e** atualiza o índice. Com `rm`, o índice continua listando arquivos que não existem. Réplicas, backups e o próprio servidor podem falhar ao ler o histórico. Se acontecer, um `PURGE BINARY LOGS TO` posterior limpa as entradas que faltam (com avisos).

E o arquivo **atual** está aberto pelo MariaDB: apagá-lo não libera espaço (o mesmo efeito do cenário 001).

## Disco cheio no MariaDB

Quando não há espaço para gravar, o MariaDB **não devolve erro**: ele espera.

```
Disk is full writing './mysql-bin.000014' (Errcode: 28). Waiting for someone to free space...
Retry in 60 secs.
```

Consequências:

- Leituras continuam funcionando; gravações ficam presas.
- Comandos que precisam do binlog (`SHOW BINARY LOGS`, `PURGE`) também ficam presos.
- Matar a consulta presa ou reiniciar o servidor pode **piorar**: desfazer uma transação (rollback) e a recuperação após o restart também precisam de espaço.

O caminho seguro é liberar espaço **no sistema de arquivos** (mover arquivos antigos para outra partição), esperar o servidor voltar sozinho (até 60 segundos) e só então fazer a limpeza com SQL.

## Arquivos que nunca se apagam na mão

| Arquivo | O que é |
|---|---|
| `ibdata1` | Espaço de tabelas do sistema InnoDB (não diminui, mesmo depois de apagar dados) |
| `ib_logfile0` | Redo log do InnoDB |
| `*.ibd`, `*.MYD`, `*.MYI` | Dados e índices das tabelas |

Apagar qualquer um deles corrompe o banco.
