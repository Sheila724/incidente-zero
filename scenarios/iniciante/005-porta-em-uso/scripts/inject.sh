#!/usr/bin/env bash
# Cenário 005 — injeta a falha:
# um servidor de testes esquecido e habilitado ganha a porta 80 no reboot.
set -euo pipefail

cat > /etc/systemd/system/debug-http.service <<'EOF'
# Servidor de teste — subir só durante a investigação do chamado #4512
[Unit]
Description=Servidor HTTP de debug (chamado 4512)
After=network.target

[Service]
ExecStart=/usr/bin/python3 -m http.server 80 --directory /
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable debug-http

# Simula o reboot: o debug chega primeiro na porta 80
systemctl stop nginx
systemctl start debug-http
for _ in $(seq 1 20); do
  ss -ltn 'sport = :80' | grep -q LISTEN && break
  sleep 0.3
done
systemctl start nginx 2>/dev/null || true

# Autoverificação: a porta 80 precisa estar expondo a listagem de arquivos
if ! curl -s http://localhost/ | grep -q "Directory listing"; then
  echo "inject: falha não reproduzida (a porta 80 não está com o debug-http)" >&2
  exit 1
fi
if systemctl is-active --quiet nginx; then
  echo "inject: o nginx subiu, a falha não foi reproduzida" >&2
  exit 1
fi
