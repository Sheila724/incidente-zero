#!/usr/bin/env bash
# Cenário 008 — prepara o ambiente SAUDÁVEL:
# API da loja guarda uma sessão por visitante em /var/lib/app/sessions;
# um timer apaga sessões com mais de 60 minutos.
set -euo pipefail

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /opt/app
install -d -o app -g app -m 0755 /var/lib/app/sessions

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: cada visita cria um arquivo de sessão."""
import json
import os
import time
import uuid
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

SESSIONS = "/var/lib/app/sessions"

class Handler(BaseHTTPRequestHandler):
    def _send(self, code, body, ctype="text/html; charset=utf-8"):
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/health":
            return self._send(200, b'{"status":"ok"}', "application/json")
        sid = uuid.uuid4().hex
        try:
            with open(os.path.join(SESSIONS, f"sess_{sid}"), "w") as f:
                json.dump({"criada": int(time.time()), "carrinho": []}, f)
        except OSError as exc:
            print(f"ERRO ao criar sessão: {exc}", flush=True)
            return self._send(503, "Serviço indisponível: não foi possível criar a sessão\n".encode())
        self._send(200, "<h1>Loja Exemplo</h1><p>Pedidos funcionando.</p>\n".encode())

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

cat > /etc/systemd/system/limpa-sessoes.service <<'EOF'
[Unit]
Description=Apaga sessões da loja com mais de 60 minutos

[Service]
Type=oneshot
ExecStart=/usr/bin/find /var/lib/app/sessions -type f -mmin +60 -delete
EOF

cat > /etc/systemd/system/limpa-sessoes.timer <<'EOF'
[Unit]
Description=Limpeza de sessões da loja a cada 5 minutos

[Timer]
OnBootSec=5min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
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
systemctl enable --now app limpa-sessoes.timer
nginx -t
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs http://localhost/ | grep -q "Loja Exemplo" && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
