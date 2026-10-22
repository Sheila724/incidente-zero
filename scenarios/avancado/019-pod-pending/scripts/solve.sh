#!/usr/bin/env bash
# Cenário 019 — solução de referência (usada pelo `iz selftest`).
# Corrige a unidade da reserva de CPU e acompanha o rollout.
set -euo pipefail
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja set resources deploy/loja-checkout --requests=cpu=250m >/dev/null
kubectl -n loja rollout status deploy/loja-checkout --timeout=120s >/dev/null
