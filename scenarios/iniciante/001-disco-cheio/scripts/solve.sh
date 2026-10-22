#!/usr/bin/env bash
# Cenário 001 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

sed -i 's/^LOG_LEVEL=.*/LOG_LEVEL=INFO/' /etc/app/app.conf
systemctl restart app
truncate -s 0 /var/log/app/debug.log
cat > /etc/logrotate.d/app <<'EOT'
/var/log/app/*.log {
    daily
    rotate 7
    maxsize 50M
    compress
    missingok
    notifempty
    copytruncate
}
EOT
