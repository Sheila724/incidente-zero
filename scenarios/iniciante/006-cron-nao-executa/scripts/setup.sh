#!/usr/bin/env bash
# Cenário 006 — prepara o ambiente SAUDÁVEL:
# dados da loja em /srv/loja/data, script de backup e o cron rodando.
set -euo pipefail

install -d -m 0755 /srv/loja/data /var/backups/loja
for i in $(seq 1 40); do
  printf '{"pedido": %d, "cliente": "cliente-%03d", "total": %d.90}\n' "$i" "$i" "$((RANDOM % 900 + 10))" \
    > "/srv/loja/data/pedido-$(printf '%04d' "$i").json"
done
echo "versao=2026.10" > /srv/loja/data/VERSION

cat > /usr/local/bin/backup-loja.sh <<'EOF'
#!/usr/bin/env bash
# Backup dos dados da loja. Mantém 7 dias de histórico.
set -euo pipefail
dest=/var/backups/loja
mkdir -p "$dest"
file="$dest/loja-$(date +%Y%m%d-%H%M).tar.gz"
tar -czf "$file" -C /srv/loja data
find "$dest" -name 'loja-*.tar.gz' -mtime +7 -delete
echo "$(date '+%F %T') backup ok: $file"
EOF
chmod 0755 /usr/local/bin/backup-loja.sh

# nginx com uma página simples, para o monitoramento HTTP
install -d -m 0755 /var/www/loja
echo '<h1>Loja Exemplo</h1><p>Servidor de dados.</p>' > /var/www/loja/index.html
cat > /etc/nginx/sites-available/loja <<'EOF'
server {
    listen 80 default_server;
    root /var/www/loja;
    index index.html;
}
EOF
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja
nginx -t
systemctl enable nginx
systemctl restart nginx

systemctl enable cron
systemctl restart cron
systemctl is-active --quiet cron
