# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. O sintoma

```bash
curl -i http://localhost
journalctl -u app -n 10 --no-pager
```

`503` e `OSError: [Errno 28] No space left on device` ao criar a sessão.

## 2. O disco tem espaço

```bash
df -h /var/lib/app
```

```
tmpfs   64M  1.2M   63M   2% /var/lib/app
```

98% livre. Mesmo assim, não dá para criar arquivos.

## 3. Os inodes acabaram

```bash
df -i /var/lib/app
```

```
tmpfs   3000  3000     0  100% /var/lib/app
```

Cada arquivo gasta um inode, e todos foram usados.

## 4. Quem gastou

```bash
sudo du --inodes -x /var/lib/app | sort -n | tail -3
ls -lt /var/lib/app/sessions | tail -3
```

Milhares de arquivos `sess_*` em `/var/lib/app/sessions`, alguns com três dias.

## 5. Libere com cuidado

Primeiro veja quantos são antigos e quantos são recentes:

```bash
sudo find /var/lib/app/sessions -type f -mmin +60 | wc -l
sudo find /var/lib/app/sessions -type f -mmin -60 | wc -l
```

Apague só os antigos:

```bash
sudo find /var/lib/app/sessions -type f -mmin +60 -delete
df -i /var/lib/app
curl -i http://localhost
```

A loja volta, e quem estava com o carrinho cheio continua logado.

Um `rm -rf /var/lib/app/sessions/*` também liberaria os inodes, mas deslogaria todo mundo. E, com mais arquivos, ele falharia com `Argument list too long`.

## 6. Por que acumulou

```bash
systemctl list-timers --all | grep -i sess
systemctl cat limpa-sessoes.timer limpa-sessoes.service
```

O timer `limpa-sessoes` apaga sessões com mais de 60 minutos a cada 5 minutos, mas foi **desativado** na migração de sexta e nunca religado.

## 7. Religue

```bash
sudo systemctl enable --now limpa-sessoes.timer
systemctl list-timers limpa-sessoes.timer
```
