#!/usr/bin/env bash
# Cenário 005 — solução de referência (usada pelo `iz selftest`).
set -euo pipefail

systemctl disable --now debug-http
systemctl start nginx
