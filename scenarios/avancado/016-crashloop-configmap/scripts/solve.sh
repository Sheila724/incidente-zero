#!/usr/bin/env bash
# Cenário 016 — solução de referência (usada pelo `iz selftest`).
# Recria a chave que a aplicação espera e recria os pods.
set -euo pipefail
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja patch configmap loja-config --type merge \
  -p '{"data":{"DATABASE_URL":"mysql://loja@db.loja.svc:3306/loja"}}' >/dev/null
kubectl -n loja rollout restart deploy/loja-web >/dev/null
kubectl -n loja rollout status deploy/loja-web --timeout=120s >/dev/null
