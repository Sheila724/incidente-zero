#!/usr/bin/env bash
# Cenário 020 — solução de referência (usada pelo `iz selftest`).
# Restaura o banco e corrige o rótulo que o colocava na economia de homologação.
set -euo pipefail
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja label deploy/loja-db ambiente=producao --overwrite >/dev/null
kubectl -n loja scale deploy/loja-db --replicas=1 >/dev/null
kubectl -n loja rollout status deploy/loja-db --timeout=120s >/dev/null
kubectl -n loja rollout status deploy/loja-web --timeout=120s >/dev/null
