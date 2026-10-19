# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Delimite o problema

```bash
curl -s http://localhost/                 # catálogo: funciona
curl -s -m 15 http://localhost/checkout   # 503 depois de alguns segundos
```

Leituras funcionam, gravações não.

## 2. Olhe o banco e o disco

```bash
df -h /var/lib/mysql
sudo journalctl -u mariadb -n 20 --no-pager
sudo mariadb -e "SHOW PROCESSLIST"
```

```
tmpfs   256M  256M  0  100% /var/lib/mysql
Disk is full writing '/var/lib/mysql/mysql-bin.000014' (Errcode: 28). Waiting for someone to free space...
INSERT INTO eventos_importados ...   Waiting for someone to free space
COMMIT                               Commit
```

A partição do banco está cheia, e o MariaDB está esperando espaço para gravar no binlog.

## 3. O que ocupa a partição

```bash
sudo ls -lh /var/lib/mysql
sudo du -ch /var/lib/mysql/mysql-bin.* | tail -1
```

Os binlogs somam a maior parte da partição: ligados para a réplica de teste, sem expiração, e inflados pela importação de eventos da madrugada.

## 4. A armadilha: tentar limpar pelo banco

```sql
SHOW BINARY LOGS;   -- fica preso
```

O comando precisa da mesma trava que a gravação presa está segurando. Reiniciar o banco ou dar `KILL` na consulta também é arriscado: desfazer a transação e a recuperação após o restart precisam de espaço, e o servidor pode cair.

## 5. Destrave: libere espaço no sistema de arquivos

Mova os binlogs **mais antigos** para outra partição. Nunca mova o último (o atual):

```bash
sudo ls /var/lib/mysql | grep mysql-bin
sudo mv /var/lib/mysql/mysql-bin.000001 /var/lib/mysql/mysql-bin.000002 /root/
df -h /var/lib/mysql
```

Em até 60 segundos o MariaDB tenta de novo e volta a gravar:

```bash
watch -n 5 "sudo mariadb -e 'SHOW PROCESSLIST'"
```

Quando sumir o `Waiting for someone to free space`, o checkout já funciona.

## 6. Limpe do jeito certo

```sql
SHOW MASTER STATUS;                          -- ex.: mysql-bin.000014
PURGE BINARY LOGS TO 'mysql-bin.000014';
SHOW WARNINGS;                               -- avisos sobre os 2 arquivos movidos: esperado
SHOW BINARY LOGS;
```

O `PURGE` apaga os arquivos e tira do índice também os que você moveu. Agora pode apagar as cópias:

```bash
sudo rm /root/mysql-bin.00000*
df -h /var/lib/mysql
```

## 7. Impeça que encha de novo

A réplica de teste não existe mais. Duas opções:

**Manter o binlog, com expiração** (útil para recuperação point-in-time):

```sql
SET GLOBAL binlog_expire_logs_seconds = 86400;
```

```bash
sudo nano /etc/mysql/mariadb.conf.d/60-loja.cnf
# binlog_expire_logs_seconds = 86400
```

**Ou desligar o binlog:** comente a linha `log_bin` no mesmo arquivo e reinicie o MariaDB (agora com espaço, o restart é seguro).

O `SET GLOBAL` vale até o próximo restart; o arquivo de configuração vale depois dele. São necessários os dois.

## Se você já apagou os binlogs com rm

```sql
FLUSH BINARY LOGS;                           -- começa um arquivo novo
SHOW MASTER STATUS;
PURGE BINARY LOGS TO 'mysql-bin.NNNNNN';     -- o arquivo atual: limpa o índice
```
