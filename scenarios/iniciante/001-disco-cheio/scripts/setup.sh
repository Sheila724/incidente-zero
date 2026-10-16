#!/usr/bin/env bash
# Cenário 001 — prepara o ambiente SAUDÁVEL:
# nginx (porta 80) -> app Python "Loja Exemplo" (127.0.0.1:8080), logs em /var/log/app
set -euo pipefail

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app

install -d -m 0755 /etc/app /opt/app
install -d -o app -g app -m 0755 /var/log/app
install -d -o www-data -g adm -m 0755 /var/log/nginx

cat > /etc/app/app.conf <<'EOF'
# Configuração da Loja Exemplo
# Níveis de log: DEBUG, INFO, WARNING, ERROR
LOG_LEVEL=INFO
PORT=8080
EOF

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: API mínima usada no cenário 001 do Incidente Zero."""
import datetime
import json
import os
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

CONF_FILE = "/etc/app/app.conf"
LOG_DIR = "/var/log/app"


def load_conf():
    conf = {"LOG_LEVEL": "INFO", "PORT": "8080"}
    try:
        with open(CONF_FILE) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                conf[key.strip()] = value.strip().strip('"')
    except FileNotFoundError:
        pass
    return conf


CONF = load_conf()
LEVEL = CONF["LOG_LEVEL"].upper()
PORT = int(CONF["PORT"])


def write_log(name, msg):
    ts = datetime.datetime.now().isoformat(timespec="seconds")
    with open(os.path.join(LOG_DIR, name), "a") as f:
        f.write(f"{ts} {msg}\n")


def debug_writer():
    """Com DEBUG ligado, despeja o cache inteiro no log o tempo todo."""
    payload = "x" * 900
    while True:
        try:
            write_log("debug.log", f"DEBUG cache_dump={payload}")
        except OSError:
            pass
        time.sleep(0.005)


class Handler(BaseHTTPRequestHandler):
    def _send(self, code, body, ctype):
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/health":
            body = json.dumps({"status": "ok", "log_level": LEVEL}).encode()
            return self._send(200, body, "application/json")
        try:
            write_log("access.log", f"INFO GET {self.path}")
        except OSError as exc:
            print(f"ERRO ao gravar log: {exc}", flush=True)
            return self._send(503, "Serviço indisponível: falha ao gravar dados\n".encode(),
                              "text/plain; charset=utf-8")
        self._send(200, "<h1>Loja Exemplo</h1><p>Pedidos funcionando.</p>\n".encode(),
                   "text/html; charset=utf-8")

    def log_message(self, *args):
        pass


if __name__ == "__main__":
    print(f"Loja Exemplo iniciando na porta {PORT} com LOG_LEVEL={LEVEL}", flush=True)
    if LEVEL == "DEBUG":
        threading.Thread(target=debug_writer, daemon=True).start()
    ThreadingHTTPServer(("127.0.0.1", PORT), Handler).serve_forever()
PY
chmod 0755 /opt/app/app.py

cat > /etc/systemd/system/app.service <<'EOF'
[Unit]
Description=Loja Exemplo (API)
After=network.target

[Service]
User=app
ExecStart=/usr/bin/python3 /opt/app/app.py
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/nginx/sites-available/loja <<'EOF'
server {
    listen 80 default_server;
    server_name _;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }
}
EOF
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

systemctl daemon-reload
systemctl enable --now app
nginx -t
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs http://localhost/ >/dev/null && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
