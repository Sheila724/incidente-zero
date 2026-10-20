#!/usr/bin/env bash
# Cenário 015 — prepara o ambiente SAUDÁVEL:
#   nginx (80) -> Loja Exemplo v2.4 (127.0.0.1:8080), com tráfego contínuo de clientes.
#   Zabbix monitora o host e a memória do processo da loja.
set -euo pipefail
umask 022

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /etc/app /opt/app /opt/trafego

cat > /etc/app/app.env <<'EOT'
# Loja Exemplo — gerenciado pelo deploy
APP_VERSION=2.4.0
EOT

cat > /opt/app/CHANGELOG.md <<'EOT'
# Loja Exemplo — changelog

## 2.4.0
- Página de promoções.
EOT

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: catálogo com sessões de clientes."""
import json
import os
import uuid
from collections import OrderedDict
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

VERSION = os.environ.get("APP_VERSION", "dev")
# Cache de sessões em memória (2.5.0). 0 = sem limite.
CACHE_MAX = int(os.environ.get("SESSION_CACHE_MAX", "-1"))
sessoes = OrderedDict()


def guardar_sessao():
    if CACHE_MAX < 0:          # versões antigas: sem cache
        return
    sid = uuid.uuid4().hex
    sessoes[sid] = {"carrinho": "x" * 50_000}   # ~50 KB por sessão
    if CACHE_MAX > 0:
        while len(sessoes) > CACHE_MAX:
            sessoes.popitem(last=False)


class H(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def send(self, code, body, ctype):
        data = body.encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        if self.path == "/health":
            return self.send(200, json.dumps({"versao": VERSION, "sessoes_em_cache": len(sessoes),
                                              "session_cache_max": CACHE_MAX}), "application/json")
        guardar_sessao()
        self.send(200, f"<h1>Loja Exemplo</h1><p>Versão {VERSION}</p>\n", "text/html; charset=utf-8")

    def log_message(self, *a):
        pass


print(f"Loja Exemplo {VERSION} (cache de sessões: {CACHE_MAX})", flush=True)
ThreadingHTTPServer(("127.0.0.1", 8080), H).serve_forever()
PY

cat > /etc/systemd/system/app.service <<'EOT'
[Unit]
Description=Loja Exemplo
After=network.target

[Service]
User=app
EnvironmentFile=/etc/app/app.env
ExecStart=/usr/bin/python3 /opt/app/app.py
Restart=always
RestartSec=2
# Proteção do servidor: acima disso o systemd encerra a loja
MemoryMax=900M

[Install]
WantedBy=multi-user.target
EOT

# Clientes navegando na loja (simulação de tráfego)
cat > /opt/trafego/clientes.py <<'PY'
#!/usr/bin/env python3
"""Simula clientes navegando: cerca de 10 páginas por segundo."""
import time
import urllib.request

while True:
    try:
        urllib.request.urlopen("http://127.0.0.1/", timeout=2).read()
    except OSError:
        pass
    time.sleep(0.1)
PY
cat > /etc/systemd/system/trafego-clientes.service <<'EOT'
[Unit]
Description=Tráfego de clientes da Loja Exemplo (simulação)
After=nginx.service

[Service]
DynamicUser=yes
ExecStart=/usr/bin/python3 /opt/trafego/clientes.py
Restart=always

[Install]
WantedBy=multi-user.target
EOT

cat > /etc/nginx/sites-available/loja <<'EOT'
server {
    listen 80 default_server;
    server_name _;
    access_log off;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }
}
EOT
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

sed -i 's/^Hostname=.*/Hostname=loja-web-01/' /etc/zabbix/zabbix_agent2.conf
systemctl daemon-reload
systemctl enable --now app
zabbix-lab-up
systemctl enable --now trafego-clientes

# ---------- Monitoramento ----------
host=$(zabbix-lab-host loja-web-01)
iface=$(zbx-api hostinterface.get "{\"hostids\":\"$host\",\"output\":[\"interfaceid\"]}" | jq -r '.[0].interfaceid')
zbx-api item.create "{\"hostid\":\"$host\",\"interfaceid\":\"$iface\",\"type\":0,
  \"name\":\"Loja Exemplo: memória do processo\",\"key_\":\"proc.mem[python3,app,,app.py,rss]\",
  \"value_type\":3,\"delay\":\"10s\",\"units\":\"B\"}" >/dev/null
zbx-api item.create "{\"hostid\":\"$host\",\"interfaceid\":\"$iface\",\"type\":0,
  \"name\":\"Loja Exemplo: porta 80\",\"key_\":\"net.tcp.service[http,,80]\",
  \"value_type\":3,\"delay\":\"10s\"}" >/dev/null
zbx-api trigger.create '[
  {"description":"Loja Exemplo usando muita memória em {HOST.NAME} ({ITEM.LASTVALUE1})",
   "expression":"last(/loja-web-01/proc.mem[python3,app,,app.py,rss])>250M","priority":3},
  {"description":"Loja Exemplo fora do ar em {HOST.NAME}",
   "expression":"max(/loja-web-01/net.tcp.service[http,,80],30s)=0","priority":5}
]' >/dev/null

for _ in $(seq 1 30); do
  curl -fs http://localhost/ | grep -q "Loja Exemplo" && exit 0
  sleep 1
done
echo "setup: a loja não respondeu" >&2
exit 1
