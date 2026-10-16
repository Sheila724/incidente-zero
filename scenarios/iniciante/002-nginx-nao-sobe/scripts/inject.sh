#!/usr/bin/env bash
# Cenário 002 — injeta a falha:
# o deploy noturno publicou a landing page da Black Friday com dois erros
# e reiniciou o nginx sem testar a configuração.
set -euo pipefail

install -d -m 0755 /var/www/promo
cat > /var/www/promo/index.html <<'EOF'
<!doctype html>
<html lang="pt-br">
<head><meta charset="utf-8"><title>Black Friday</title></head>
<body><h1>Black Friday da Loja Exemplo</h1><p>Até 70% de desconto!</p></body>
</html>
EOF

# Erro 1: falta o ";" no listen.  Erro 2: segundo default_server na porta 80.
cat > /etc/nginx/sites-available/promo <<'EOF'
# Landing page Black Friday — publicada no deploy noturno
server {
    listen 80 default_server
    server_name promo.loja.local;

    root /var/www/promo;
    index index.html;
}
EOF
ln -sf /etc/nginx/sites-available/promo /etc/nginx/sites-enabled/promo

# O deploy reiniciou sem rodar "nginx -t"
systemctl restart nginx 2>/dev/null || true

# Autoverificação: o site precisa estar fora do ar
code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/ || true)
if [ "$code" != "000" ]; then
  echo "inject: falha não reproduzida (HTTP $code)" >&2
  exit 1
fi
