#!/usr/bin/env bash
# Cenário 017 — solução de referência (usada pelo `iz selftest`).
# Aponta o Deployment para a tag publicada e acompanha o rollout.
set -euo pipefail
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja set image deploy/loja-web web=registry.loja.local/loja-web:2.6 >/dev/null
kubectl -n loja rollout status deploy/loja-web --timeout=120s >/dev/null
