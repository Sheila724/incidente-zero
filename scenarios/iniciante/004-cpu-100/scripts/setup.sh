#!/usr/bin/env bash
# Cenário 004 — prepara o ambiente SAUDÁVEL:
# nginx -> API da loja (Python, 127.0.0.1:8080) + gerador de relatórios a cada 5 min.
set -euo pipefail

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /opt/app /opt/report-worker /etc/report-worker
install -d -o app -g app -m 0755 /var/lib/report-worker

# ---------- API da loja ----------
cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: API mínima. Cada pedido faz um pouco de cálculo, como uma API real."""
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

def calcula_frete(n=20000):
    t = 0
    for i in range(n):
        t += (i * 7) % 13
    return t

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            body = json.dumps({"status": "ok"}).encode()
            ctype = "application/json"
        else:
            calcula_frete()
            body = "<h1>Loja Exemplo</h1><p>Pedidos funcionando.</p>\n".encode()
            ctype = "text/html; charset=utf-8"
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass

ThreadingHTTPServer(("127.0.0.1", 8080), Handler).serve_forever()
PY

cat > /etc/systemd/system/app.service <<'EOF'
[Unit]
Description=Loja Exemplo (API)
After=network.target

[Service]
User=app
ExecStart=/usr/bin/python3 /opt/app/app.py
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# ---------- Gerador de relatórios ----------
cat > /opt/report-worker/worker.py <<'PY'
#!/usr/bin/env python3
"""Gerador de relatórios de vendas: gera um relatório e espera INTERVAL_SECONDS."""
import datetime
import json
import os
import random
import time

CONF = "/etc/report-worker/worker.conf"
OUT = "/var/lib/report-worker/report-latest.json"

def load_interval():
    try:
        with open(CONF) as f:
            for line in f:
                line = line.strip()
                if line.startswith("INTERVAL_SECONDS="):
                    return int(line.split("=", 1)[1])
    except (OSError, ValueError):
        pass
    return 300

interval = load_interval()
print(f"report-worker iniciado: intervalo de {interval}s", flush=True)
n = 0
while True:
    vendas = [random.uniform(10, 900) for _ in range(3000)]
    relatorio = {
        "gerado_em": datetime.datetime.now().isoformat(timespec="seconds"),
        "pedidos": len(vendas),
        "faturamento": round(sum(vendas), 2),
    }
    tmp = OUT + ".tmp"
    with open(tmp, "w") as f:
        json.dump(relatorio, f)
    os.replace(tmp, OUT)
    n += 1
    if interval > 0 or n % 5000 == 0:
        print(f"relatório #{n} gerado: {relatorio['pedidos']} pedidos", flush=True)
    time.sleep(interval)
PY

cat > /etc/report-worker/worker.conf <<'EOF'
# Gerador de relatórios de vendas
# Intervalo entre relatórios, em segundos
INTERVAL_SECONDS=300
EOF

cat > /etc/systemd/system/report-worker.service <<'EOF'
[Unit]
Description=Gerador de relatórios de vendas
After=network.target

[Service]
User=app
ExecStart=/usr/bin/python3 /opt/report-worker/worker.py
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF

# ---------- nginx ----------
cat > /etc/nginx/sites-available/loja <<'EOF'
server {
    listen 80 default_server;
    server_name _;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_read_timeout 10s;
    }
}
EOF
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

systemctl daemon-reload
systemctl enable --now app report-worker
nginx -t
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs http://localhost/ | grep -q "Loja Exemplo" && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
