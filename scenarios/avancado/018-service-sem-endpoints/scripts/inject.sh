#!/usr/bin/env bash
# Cenário 018 — injeta a falha:
# a "padronização das health checks" trocou a readiness probe para /healthz,
# mas a loja só atende /health. Com Recreate, nenhum pod fica pronto.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja patch deploy loja-web --type json -p '[
  {"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/healthz"},
  {"op":"add","path":"/metadata/annotations/plataforma.loja.local~1padronizacao","value":"health-checks v1: readinessProbe em /healthz"}
]' >/dev/null

# Autoverificação: pods rodando, nenhum pronto, Service sem endpoints e 502
for _ in $(seq 1 60); do
  rodando=$(kubectl -n loja get pods -l app=loja-web --field-selector=status.phase=Running -o name 2>/dev/null | wc -l)
  prontos=$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
  falhas=$(kubectl -n loja get events --field-selector reason=Unhealthy -o name 2>/dev/null | wc -l)
  code=$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)
  if [ "$rodando" -ge 2 ] && [ "${prontos:-0}" = "0" ] && [ "$falhas" -ge 1 ] && [ "$code" = "502" ]; then
    exit 0
  fi
  sleep 2
done
echo "inject: a loja não ficou sem endpoints como esperado" >&2
exit 1
