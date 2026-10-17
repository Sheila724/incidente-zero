# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Confirme o sintoma

```bash
ls -lt /var/backups/loja
```

O mais recente é de três dias atrás.

## 2. O script funciona?

```bash
sudo /usr/local/bin/backup-loja.sh
ls -lt /var/backups/loja | head -3
```

Funciona: gerou um backup novo. O problema não é o script, é o agendamento.

## 3. Onde está o agendamento

```bash
ls -l /etc/cron.d/
cat /etc/cron.d/backup-loja.cron
```

```
*/2 * * * * root /usr/local/bin/backup-loja.sh > /var/log/backup-loja-$(date +%F).log 2>&1
```

## 4. O cron está executando?

```bash
journalctl -u cron --since "-10 min" --no-pager
```

Nenhuma linha com `backup-loja`. O cron nem está lendo esse arquivo.

## 5. Primeiro erro: o nome

O cron ignora arquivos do `/etc/cron.d` com ponto no nome. A "padronização" de segunda-feira adicionou o `.cron`.

```bash
sudo mv /etc/cron.d/backup-loja.cron /etc/cron.d/backup-loja
```

## 6. Segundo erro: o `%`

Dois minutos depois, o journal mostra a execução, mas nenhum backup novo aparece:

```bash
journalctl -u cron --since "-5 min" --no-pager
```

```
CRON[311]: (root) CMD (/usr/local/bin/backup-loja.sh > /var/log/backup-loja-$(date +)
CRON[310]: (CRON) info (No MTA installed, discarding output)
```

O comando foi cortado no `%`. O `sh` recebeu `$(date +` e deu erro de sintaxe, e a mensagem de erro foi descartada por falta de servidor de e-mail.

Corrija escapando o `%`:

```bash
sudo sed -i 's/%F/\\%F/' /etc/cron.d/backup-loja
cat /etc/cron.d/backup-loja
```

## 7. Valide

Espere até 2 minutos:

```bash
ls -lt /var/backups/loja | head -3
tar -tzf "$(ls -t /var/backups/loja/*.tar.gz | head -1)" | head
cat /var/log/backup-loja-*.log
```

Backup novo, conteúdo válido e log com `backup ok`.
