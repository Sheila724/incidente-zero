#!/usr/bin/env bash
# Cenário 008 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

find /var/lib/app/sessions -type f -mmin +60 -delete
systemctl enable --now limpa-sessoes.timer
