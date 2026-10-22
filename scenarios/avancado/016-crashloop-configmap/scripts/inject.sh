#!/usr/bin/env bash
# Cenário 016 — injeta a falha:
# ontem a ConfigMap foi "padronizada" (DATABASE_URL -> DB_URL), sem rollout.
# Na manutenção das 03:00, os pods foram recriados e a aplicação não sobe mais.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-config
  annotations:
    loja.local/alterado-em: "ontem, 18:40"
    loja.local/motivo: "Padronização dos nomes das variáveis"
data:
  DB_URL: mysql://loja@db.loja.svc:3306/loja
  CACHE_TTL: "300"
YAML

# Manutenção do cluster: todos os pods recriados
kubectl -n loja delete pod -l app=loja-web --wait=false >/dev/null

# Autoverificação: pods reiniciando e a loja fora do ar
for _ in $(seq 1 60); do
  r=$(kubectl -n loja get pods -l app=loja-web -o jsonpath='{range .items[*]}{.status.containerStatuses[0].restartCount}{"\n"}{end}' 2>/dev/null | sort -n | head -1)
  code=$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)
  [ "${r:-0}" -ge 1 ] && [ "$code" = "502" ] && exit 0
  sleep 2
done
echo "inject: os pods não entraram em falha" >&2
exit 1
