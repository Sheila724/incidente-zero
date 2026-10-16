# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Reproduza o sintoma

```bash
curl -i http://localhost
```

Resposta `503` com *"falha ao gravar dados"*. O nginx está de pé e repassou o erro da aplicação: o problema é **atrás** dele.

## 2. Olhe o serviço

```bash
systemctl status app
journalctl -u app -n 20
```

O serviço está ativo, mas o journal mostra `ERRO ao gravar log: [Errno 28] No space left on device`. O erro 28 (`ENOSPC`) significa disco cheio.

## 3. Ache a partição cheia

```bash
df -h
```

`/var/log` aparece com 100% de uso. As outras partições estão normais.

## 4. Ache o que ocupa espaço

```bash
sudo du -sh /var/log/* | sort -h
sudo du -sh /var/log/app/* | sort -h
```

`/var/log/app/debug.log` ocupa praticamente toda a partição.

## 5. Entenda por que ele cresce

```bash
sudo tail -n 3 /var/log/app/debug.log
cat /etc/app/app.conf
```

> ⚠️ Não use `cat` no arquivo de log gigante: ele vai inundar seu terminal. Use `tail`, `head` ou `less`.

O log está cheio de linhas `DEBUG cache_dump=...`, e o arquivo de configuração tem `LOG_LEVEL=DEBUG` com um comentário dizendo que era temporário. **Essa é a causa raiz.** Se você só apagar o arquivo, ele volta a encher.

## 6. Corrija a causa

```bash
sudo sed -i 's/^LOG_LEVEL=.*/LOG_LEVEL=INFO/' /etc/app/app.conf
sudo systemctl restart app
curl -s http://127.0.0.1:8080/health
```

O `/health` deve mostrar `"log_level": "INFO"`. Isso confirma que a aplicação **em execução** pegou a mudança.

## 7. Libere o espaço

```bash
sudo truncate -s 0 /var/log/app/debug.log
df -h /var/log
```

Em produção, antes de zerar, você poderia guardar um trecho para análise:

```bash
sudo tail -n 1000 /var/log/app/debug.log > ~/debug-incidente.log
```

## 8. Valide

```bash
curl -i http://localhost
```

`HTTP/1.1 200 OK`. Incidente mitigado.

## 9. Bônus: prevenção

```bash
sudo tee /etc/logrotate.d/app > /dev/null <<'EOF'
/var/log/app/*.log {
    daily
    rotate 7
    maxsize 50M
    compress
    missingok
    notifempty
    copytruncate
}
EOF
sudo logrotate -d /etc/logrotate.d/app   # simula, sem aplicar
```

`copytruncate` copia e zera o arquivo original, o que funciona com aplicações que mantêm o log aberto.
