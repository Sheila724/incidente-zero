#!/usr/bin/env bash
# Cenário 008 — injeta a falha:
# o timer de limpeza foi pausado na migração e as sessões esgotaram os inodes.
set -euo pipefail

systemctl disable --now limpa-sessoes.timer 2>/dev/null
sed -i '1i # Pausado durante a migração de sexta (OPS-77) — religar depois da janela' \
  /etc/systemd/system/limpa-sessoes.timer
systemctl daemon-reload

install -d -m 0700 /var/lib/incidente-zero
python3 - <<'PY'
import json, os, random, time, uuid

d = "/var/lib/app/sessions"
uid = int(os.popen("id -u app").read())
gid = int(os.popen("id -g app").read())
names = []
while True:
    name = os.path.join(d, f"sess_{uuid.uuid4().hex}")
    try:
        with open(name, "w") as f:
            json.dump({"criada": 0, "carrinho": []}, f)
    except OSError:
        break
    names.append(name)

now = time.time()
recent = names[-40:]
for n in names:
    if n in recent:
        t = now - random.randint(60, 30 * 60)           # últimos 30 minutos
    else:
        t = now - random.randint(2 * 3600, 72 * 3600)   # entre 2 h e 3 dias
    os.utime(n, (t, t))
    os.chown(n, uid, gid)

with open("/var/lib/incidente-zero/sessoes-recentes.txt", "w") as f:
    f.write("\n".join(recent) + "\n")
print(f"{len(names)} sessões criadas")
PY

# Autoverificação: inodes esgotados, espaço sobrando, loja em 503
ipct=$(df --output=ipcent /var/lib/app | tail -1 | tr -dc 0-9)
spct=$(df --output=pcent /var/lib/app | tail -1 | tr -dc 0-9)
code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/ || true)
if [ "$ipct" -lt 99 ] || [ "$spct" -gt 50 ] || [ "$code" != "503" ]; then
  echo "inject: falha não reproduzida (inodes ${ipct}%, espaço ${spct}%, HTTP ${code})" >&2
  exit 1
fi
