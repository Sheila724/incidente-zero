#!/usr/bin/env bash
# Cenário 002 — prepara o ambiente SAUDÁVEL:
# nginx servindo o site estático da Loja Exemplo como site padrão.
set -euo pipefail

install -d -m 0755 /var/www/loja
cat > /var/www/loja/index.html <<'EOF'
<!doctype html>
<html lang="pt-br">
<head><meta charset="utf-8"><title>Loja Exemplo</title></head>
<body><h1>Loja Exemplo</h1><p>Catálogo e pedidos funcionando.</p></body>
</html>
EOF

cat > /etc/nginx/sites-available/loja <<'EOF'
# Site principal da Loja Exemplo
server {
    listen 80 default_server;
    server_name loja.local;

    root /var/www/loja;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
EOF

rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

nginx -t
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs http://localhost/ | grep -q "Loja Exemplo" && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
