# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Confirme o sintoma

```bash
uptime
top
```

O load average está alto e, no `top`, um processo `python3` aparece usando quase toda a CPU disponível.

## 2. Identifique o processo

```bash
ps aux --sort=-%cpu | head -5
```

O comando é `/usr/bin/python3 /opt/report-worker/worker.py`. Anote o PID.

## 3. Descubra o dono

```bash
systemctl status <PID>
```

O processo pertence ao serviço `report-worker.service`.

## 4. A armadilha: matar não resolve

```bash
sudo kill <PID>
sleep 3
ps aux --sort=-%cpu | head -3
```

O worker voltou, com outro PID. O motivo:

```bash
systemctl cat report-worker
```

```ini
[Service]
ExecStart=/usr/bin/python3 /opt/report-worker/worker.py
Restart=always
```

## 5. Ache a causa

```bash
journalctl -u report-worker -n 20 --no-pager
cat /etc/report-worker/worker.conf
```

O log mostra relatórios sendo gerados sem parar, e a configuração diz:

```ini
# Ajuste do deploy das 15:30 — testar geração contínua
INTERVAL_SECONDS=0
```

O worker gera um relatório e dorme `INTERVAL_SECONDS` antes do próximo. Com `0`, ele não dorme nunca: vira um loop que ocupa toda a CPU.

## 6. Corrija e reinicie

```bash
sudo sed -i 's/^INTERVAL_SECONDS=.*/INTERVAL_SECONDS=300/' /etc/report-worker/worker.conf
sudo systemctl restart report-worker
top
```

A CPU volta ao normal e o serviço continua ativo, gerando relatórios a cada 5 minutos.

## 7. Bônus: um limite de segurança

```bash
sudo systemctl edit report-worker
```

```ini
[Service]
CPUQuota=20%
```

```bash
sudo systemctl restart report-worker
systemctl show -p CPUQuotaPerSecUSec report-worker
```

Mesmo que alguém repita o erro, o worker nunca mais vai passar de 20% de um núcleo e a loja continua respondendo.
