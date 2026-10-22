#!/usr/bin/env bash
# Cenário 018 — solução de referência (usada pelo `iz selftest`).
# Aponta a readiness probe para o caminho que a aplicação atende.
set -euo pipefail
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja patch deploy loja-web --type json -p \
  '[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/health"}]' >/dev/null
kubectl -n loja rollout status deploy/loja-web --timeout=120s >/dev/null
