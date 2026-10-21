#!/usr/bin/env bash
# Cenário 004 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

sed -i 's/^INTERVAL_SECONDS=.*/INTERVAL_SECONDS=300/' /etc/report-worker/worker.conf
mkdir -p /etc/systemd/system/report-worker.service.d
printf '[Service]\nCPUQuota=20%%\n' > /etc/systemd/system/report-worker.service.d/limite.conf
systemctl daemon-reload
systemctl restart report-worker
