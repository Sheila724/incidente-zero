#!/usr/bin/env bash
# Cenário 003 — prepara o ambiente SAUDÁVEL:
# nginx servindo a Loja Exemplo com home, página de produtos e imagens.
set -euo pipefail

id deploy &>/dev/null || useradd -m -s /bin/bash deploy

install -d -m 0755 /var/www/loja /var/www/loja/produtos /var/www/loja/img

cat > /var/www/loja/index.html <<'EOF'
<!doctype html>
<html lang="pt-br">
<head><meta charset="utf-8"><title>Loja Exemplo</title></head>
<body><img src="/img/logo.svg" alt="logo"><h1>Loja Exemplo</h1>
<p><a href="/produtos/">Ver produtos</a></p></body>
</html>
EOF

cat > /var/www/loja/produtos/index.html <<'EOF'
<!doctype html>
<html lang="pt-br">
<head><meta charset="utf-8"><title>Produtos</title></head>
<body><h1>Produtos</h1><ul><li>Teclado mecânico</li><li>Monitor 27"</li></ul></body>
</html>
EOF

cat > /var/www/loja/img/logo.svg <<'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64"><circle cx="32" cy="32" r="28" fill="#e4572e"/></svg>
EOF

chown -R deploy:deploy /var/www/loja
find /var/www/loja -type d -exec chmod 0755 {} +
find /var/www/loja -type f -exec chmod 0644 {} +

cat > /etc/nginx/sites-available/loja <<'EOF'
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
  curl -fs http://localhost/ | grep -q "<h1>Loja Exemplo</h1>" && exit 0
  sleep 0.5
done
echo "setup: a Loja Exemplo não respondeu" >&2
exit 1
