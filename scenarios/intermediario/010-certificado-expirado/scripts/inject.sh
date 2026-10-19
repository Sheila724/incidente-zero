#!/usr/bin/env bash
# Cenário 010 — injeta a falha:
# o servidor foi migrado para um template novo, e o timer de renovação
# não foi habilitado. O certificado venceu ontem.
set -euo pipefail
umask 022

systemctl disable --now renova-certificado.timer 2>/dev/null

# Emite (com datas no passado) o certificado que estava em uso: venceu ontem
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/novos"
: > "$tmp/index.txt"
echo 1000 > "$tmp/serial"
cat > "$tmp/ca.cnf" <<EOT
[ca]
default_ca = loja
[loja]
dir           = $tmp
database      = $tmp/index.txt
new_certs_dir = $tmp/novos
serial        = $tmp/serial
certificate   = /etc/ssl/loja-ca/ca.crt
private_key   = /etc/ssl/loja-ca/ca.key
default_md    = sha256
policy        = livre
copy_extensions = copy
unique_subject = no
[livre]
commonName = supplied
[ext]
subjectAltName = DNS:loja.local,DNS:www.loja.local
EOT
inicio=$(date -u -d '91 days ago' +%Y%m%d%H%M%SZ)
fim=$(date -u -d '1 day ago' +%Y%m%d%H%M%SZ)
openssl req -new -key /etc/ssl/loja/loja.key -subj "/CN=loja.local" -out "$tmp/loja.csr"
openssl ca -batch -config "$tmp/ca.cnf" -extensions ext -notext \
  -startdate "$inicio" -enddate "$fim" -in "$tmp/loja.csr" -out "$tmp/loja.crt" 2>/dev/null
install -m 0644 "$tmp/loja.crt" /etc/ssl/loja/loja.crt
touch -d '91 days ago' /etc/ssl/loja/loja.crt
systemctl reload nginx

# Autoverificação: o HTTPS precisa estar falhando por certificado vencido
sleep 1
if curl -fs -o /dev/null --resolve loja.local:443:127.0.0.1 https://loja.local/; then
  echo "inject: falha não reproduzida (o HTTPS ainda é aceito)" >&2
  exit 1
fi
openssl s_client -connect 127.0.0.1:443 -servername loja.local </dev/null 2>/dev/null \
  | openssl x509 -checkend 0 -noout >/dev/null && { echo "inject: o certificado servido não está vencido" >&2; exit 1; }
exit 0
