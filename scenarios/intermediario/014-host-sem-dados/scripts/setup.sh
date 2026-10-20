#!/usr/bin/env bash
# Cenário 014 — prepara o ambiente SAUDÁVEL:
#   Zabbix 7.0 (server, frontend em :8081 e agente 2) monitorando o host loja-web-01:
#   disponibilidade do agente, uso de /var/log e carga.
#   nginx serve a Loja Exemplo na porta 80.
set -euo pipefail
umask 022

# ---------- Loja Exemplo (site estático) ----------
install -d -m 0755 /var/www/loja
echo '<h1>Loja Exemplo</h1><p>Catálogo funcionando.</p>' > /var/www/loja/index.html
cat > /etc/nginx/sites-available/loja <<'EOT'
server {
    listen 80 default_server;
    server_name _;
    root /var/www/loja;
}
EOT
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

# ---------- Agente ----------
sed -i 's/^Hostname=.*/Hostname=loja-web-01/' /etc/zabbix/zabbix_agent2.conf

zabbix-lab-up

# ---------- Monitoramento do host ----------
zabbix-lab-host loja-web-01 >/dev/null
