#!/usr/bin/env bash
# Cenário 003 — solução de referência (usada pelo `iz selftest`).
# O nginx entra no grupo dos arquivos; ninguém mais tem acesso.
set -euo pipefail

sed -i -e 's/^chown -R deploy:deploy/chown -R deploy:www-data/' \
       -e 's/^chmod -R o-rwx/chmod -R u=rwX,g=rX,o=/' /usr/local/bin/hardening-web.sh
systemctl start hardening-web.service
