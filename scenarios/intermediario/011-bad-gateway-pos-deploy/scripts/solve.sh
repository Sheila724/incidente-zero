#!/usr/bin/env bash
# Cenário 011 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

sed -i 's/server 127\.0\.0\.1:8080;/server 127.0.0.1:8081;/' /etc/nginx/conf.d/upstream-loja.conf
sed -i 's#proxy_pass http://127\.0\.0\.1:8080;#proxy_pass http://loja_app;#' /etc/nginx/sites-available/loja
nginx -t
systemctl reload nginx
