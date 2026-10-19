# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Confirme o sintoma

```bash
curl -s http://localhost/
sudo journalctl -u app -n 5 --no-pager
```

```
ERRO banco: (1040, 'Too many connections')
```

## 2. Veja quem ocupa o banco

```bash
sudo mariadb
```

```sql
SHOW VARIABLES LIKE 'max_connections';        -- 30
SELECT user, command, COUNT(*) AS conexoes
  FROM information_schema.processlist
 GROUP BY user, command;
```

```
relatorio   Sleep   30
root        Query    1
```

As 30 conexões são do usuário `relatorio`, todas paradas.

## 3. Ache o processo

```bash
ps -ef | grep -v grep | grep relatorio
cat /etc/cron.d/relatorio-vendas
```

Várias execuções de `/opt/relatorios/relatorio-vendas.py`, uma por minuto, todas vivas.

```bash
less /opt/relatorios/relatorio-vendas.py
```

A v2 abre **uma conexão por categoria** (5) e, no fim, espera até 10 minutos pelo upload do FTP **sem fechar nenhuma**. Com o cron a cada minuto: 6 execuções × 5 conexões = 30, o limite do banco.

## 4. A armadilha: reiniciar o banco

```bash
sudo systemctl restart mariadb
```

A loja volta... e, em poucos minutos, o cron enche o banco de novo. O restart também derrubou todas as conexões da loja, ou seja, mais um pico de erros para os clientes.

## 5. Corrija o vazamento

Feche as conexões assim que o arquivo estiver pronto, antes da espera:

```bash
sudo nano /opt/relatorios/relatorio-vendas.py
```

```python
# depois de gravar o CSV, antes do time.sleep(600):
for conn in conexoes.values():
    conn.close()
```

(Melhor ainda: uma única conexão, aberta com `with`, para as cinco consultas. Paralelizar cinco `COUNT(*)` em tabelas pequenas não ganha nada.)

## 6. Libere as conexões presas

As execuções antigas continuam rodando com o código velho:

```bash
sudo pkill -u relatorios -f relatorio-vendas
sudo mariadb -e "SELECT user, COUNT(*) FROM information_schema.processlist GROUP BY user"
curl -s http://localhost/checkout
```

`pedido confirmado`. Sem restart do banco.

## 7. Proteja o banco de um próximo vazamento

```sql
ALTER USER 'relatorio'@'127.0.0.1' WITH MAX_USER_CONNECTIONS 5;
```

Agora, mesmo que um relatório volte a vazar, ele só consegue ocupar 5 conexões. A loja continua com as outras 25.
