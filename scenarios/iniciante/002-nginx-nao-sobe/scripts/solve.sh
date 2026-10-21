#!/usr/bin/env bash
# Cenário 002 — solução de referência (usada pelo `iz selftest`).
# Corrige a landing page: ";" no listen e sem um segundo default_server.
set -euo pipefail

sed -i 's/^\(\s*\)listen 80 default_server$/\1listen 80;/' /etc/nginx/sites-available/promo
nginx -t
systemctl start nginx
