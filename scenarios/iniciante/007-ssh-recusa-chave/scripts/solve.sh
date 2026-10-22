#!/usr/bin/env bash
# Cenário 007 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

find /home/deploy -type d -exec chmod 755 {} +
find /home/deploy -type f -exec chmod 644 {} +
chown -R deploy:deploy /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
chmod 755 /home/deploy/app/deploy.sh
