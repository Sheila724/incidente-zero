#!/usr/bin/env bash
# Cenário 011 — prepara o ambiente SAUDÁVEL:
#   nginx (80) como proxy reverso -> Loja Exemplo v2.3 (127.0.0.1:8080)
set -euo pipefail

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /etc/app /opt/app

cat > /etc/app/app.env <<'EOT'
# Loja Exemplo — variáveis de ambiente (gerenciado pelo deploy)
APP_VERSION=2.3.0
PORT=8080
EOT

cat > /opt/app/CHANGELOG.md <<'EOT'
# Loja Exemplo — changelog

## 2.3.0
- Catálogo com filtros por categoria.
EOT

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: site e API de produtos atrás do nginx."""
import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

VERSION = os.environ.get("APP_VERSION", "dev")
PORT = int(os.environ.get("PORT", "8080"))
PRODUTOS = [{"id": 1, "nome": "Camiseta", "preco": 59.9},
            {"id": 2, "nome": "Caneca", "preco": 34.9}]


class H(BaseHTTPRequestHandler):
    def send(self, code, body, ctype):
        data = body.encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.send_header("X-App-Version", VERSION)
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        if self.path.startswith("/api/produtos"):
            return self.send(200, json.dumps({"versao": VERSION, "produtos": PRODUTOS}),
                             "application/json")
        if self.path == "/health":
            return self.send(200, json.dumps({"status": "ok", "versao": VERSION, "porta": PORT}),
                             "application/json")
        self.send(200, f"<h1>Loja Exemplo</h1><p>Versão {VERSION}</p>\n", "text/html; charset=utf-8")

    def log_message(self, *a):
        pass


print(f"Loja Exemplo {VERSION} escutando em 127.0.0.1:{PORT}", flush=True)
ThreadingHTTPServer(("127.0.0.1", PORT), H).serve_forever()
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

[Install]
WantedBy=multi-user.target
EOT

cat > /etc/nginx/conf.d/upstream-loja.conf <<'EOT'
# Backends da Loja Exemplo
upstream loja_app {
    server 127.0.0.1:8080;
    keepalive 16;
}
EOT

cat > /etc/nginx/sites-available/loja <<'EOT'
server {
    listen 80 default_server;
    server_name _;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;

    location / {
        proxy_pass http://loja_app;
    }

    # API: timeout menor para o app mobile (ajuste do incidente INC-1877)
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_read_timeout 5s;
    }
}
EOT
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

systemctl daemon-reload
systemctl enable --now app
nginx -t 2>/dev/null
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs http://localhost/api/produtos | grep -q '"2.3.0"' && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
