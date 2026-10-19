#!/usr/bin/env bash
# Cenário 010 — prepara o ambiente SAUDÁVEL:
#   nginx servindo https://loja.local com certificado emitido pela CA interna,
#   renovado automaticamente por um timer do systemd.
set -euo pipefail
umask 022

install -d -m 0755 /var/www/loja /etc/ssl/loja
install -d -m 0700 /etc/ssl/loja-ca

cat > /var/www/loja/index.html <<'EOT'
<!doctype html>
<html lang="pt-br">
<head><meta charset="utf-8"><title>Loja Exemplo</title></head>
<body><h1>Loja Exemplo</h1><p>Conexão segura funcionando.</p></body>
</html>
EOT

# ---------- CA interna ----------
openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 3650 \
  -subj "/O=Loja Exemplo/CN=CA Interna Loja Exemplo" \
  -keyout /etc/ssl/loja-ca/ca.key -out /etc/ssl/loja-ca/ca.crt 2>/dev/null
cp /etc/ssl/loja-ca/ca.crt /usr/local/share/ca-certificates/ca-interna-loja.crt
update-ca-certificates >/dev/null 2>&1

openssl genrsa -out /etc/ssl/loja/loja.key 2048 2>/dev/null
chmod 0600 /etc/ssl/loja/loja.key

# ---------- renovação automática ----------
cat > /usr/local/sbin/renova-certificado <<'EOT'
#!/usr/bin/env bash
# Renova o certificado de loja.local pela CA interna
# quando faltam menos de 30 dias para ele vencer.
set -euo pipefail

CRT=/etc/ssl/loja/loja.crt
KEY=/etc/ssl/loja/loja.key
CA=/etc/ssl/loja-ca

if openssl x509 -checkend $((30 * 86400)) -noout -in "$CRT" >/dev/null 2>&1; then
  echo "Certificado válido por mais de 30 dias: nada a fazer."
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
openssl req -new -key "$KEY" -subj "/CN=loja.local" -out "$tmp/loja.csr"
printf 'subjectAltName=DNS:loja.local,DNS:www.loja.local\n' > "$tmp/ext.cnf"
openssl x509 -req -in "$tmp/loja.csr" -CA "$CA/ca.crt" -CAkey "$CA/ca.key" \
  -CAcreateserial -days 90 -sha256 -extfile "$tmp/ext.cnf" -out "$tmp/loja.crt" 2>/dev/null
install -m 0644 "$tmp/loja.crt" "$CRT"

echo "Certificado renovado. Válido até: $(openssl x509 -enddate -noout -in "$CRT" | cut -d= -f2)"
EOT
chmod 0755 /usr/local/sbin/renova-certificado

cat > /etc/systemd/system/renova-certificado.service <<'EOT'
[Unit]
Description=Renovação do certificado TLS de loja.local

[Service]
Type=oneshot
ExecStart=/usr/local/sbin/renova-certificado
EOT

cat > /etc/systemd/system/renova-certificado.timer <<'EOT'
[Unit]
Description=Verifica a validade do certificado de loja.local todo dia

[Timer]
OnCalendar=daily
RandomizedDelaySec=1h
Persistent=true

[Install]
WantedBy=timers.target
EOT

# Primeira emissão
/usr/local/sbin/renova-certificado >/dev/null

# ---------- nginx ----------
cat > /etc/nginx/sites-available/loja <<'EOT'
# Loja Exemplo — HTTPS
server {
    listen 443 ssl default_server;
    server_name loja.local www.loja.local;

    ssl_certificate     /etc/ssl/loja/loja.crt;
    ssl_certificate_key /etc/ssl/loja/loja.key;

    root /var/www/loja;
    index index.html;
}

# Porta 80: só para o health check interno
server {
    listen 80 default_server;
    server_name _;
    root /var/www/loja;
    index index.html;
}
EOT
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

systemctl daemon-reload
systemctl enable --now renova-certificado.timer
nginx -t 2>/dev/null
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 20); do
  curl -fs --resolve loja.local:443:127.0.0.1 https://loja.local/ | grep -q "Loja Exemplo" && exit 0
  sleep 0.5
done
echo "setup: https://loja.local não respondeu" >&2
exit 1
