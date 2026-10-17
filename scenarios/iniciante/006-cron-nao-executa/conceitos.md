# Conceitos para este cenário

## A linha do cron

```
┌──────── minuto (0-59)
│ ┌────── hora (0-23)
│ │ ┌──── dia do mês (1-31)
│ │ │ ┌── mês (1-12)
│ │ │ │ ┌ dia da semana (0-7, 0 e 7 = domingo)
│ │ │ │ │
*/2 * * * *  root  /usr/local/bin/backup-loja.sh
             └ usuário (só em /etc/cron.d e /etc/crontab)
```

| Expressão | Significado |
|---|---|
| `0 2 * * *` | Todo dia às 02:00 |
| `*/2 * * * *` | A cada 2 minutos |
| `30 8 * * 1-5` | 08:30, de segunda a sexta |

O site [crontab.guru](https://crontab.guru) traduz qualquer expressão.

## Onde ficam os agendamentos

| Local | Quem edita | Tem coluna de usuário? |
|---|---|---|
| `crontab -e` (por usuário) | o próprio usuário | Não |
| `/etc/crontab` | root | Sim |
| `/etc/cron.d/*` | root, pacotes | Sim |
| `/etc/cron.daily/` etc. | root | Não (são scripts) |

## A regra de nomes do `/etc/cron.d`

No Debian e no Ubuntu, o cron **ignora em silêncio** arquivos do `/etc/cron.d` cujo nome tenha caracteres fora de letras, números, `_` e `-`. Ou seja: `backup-loja` funciona, `backup-loja.cron` e `backup.sh` são ignorados. Nenhum erro aparece em lugar nenhum.

A regra existe para não executar arquivos como `algo.dpkg-old` ou `algo~`, que sobram de atualizações e editores.

## O `%` na linha do cron

Dentro do comando de uma linha do cron, `%` vira **quebra de linha**, e o que vem depois vai para a entrada padrão do comando. Então:

```
* * * * * root script.sh > /var/log/b-$(date +%F).log
```

é executado como `script.sh > /var/log/b-$(date +`, que é um erro de sintaxe. Para usar `%` de verdade, escape: `\%F`. Ou, melhor, coloque a lógica dentro do script.

## O ambiente do cron é mínimo

O cron roda com `PATH=/usr/bin:/bin` e quase nenhuma variável de ambiente. Use caminhos absolutos para comandos e scripts.

## Confirmando execuções

```bash
journalctl -u cron --since "-15 min"
```

Cada execução aparece como `(root) CMD (...)`. Se o agendamento não aparece no journal, o cron nem leu a linha.

## Backup não é backup até ser restaurado

Um arquivo `.tar.gz` existir não prova nada. No mínimo, liste o conteúdo:

```bash
tar -tzf /var/backups/loja/loja-20261016-0915.tar.gz | head
```
