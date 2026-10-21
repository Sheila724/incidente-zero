#!/usr/bin/env bash
# Cenário 010 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

mkdir -p /etc/systemd/system/renova-certificado.service.d
printf '[Service]\nExecStartPost=/usr/bin/systemctl reload nginx\n' \
  > /etc/systemd/system/renova-certificado.service.d/reload-nginx.conf
systemctl daemon-reload
systemctl enable --now renova-certificado.timer
systemctl start renova-certificado.service
