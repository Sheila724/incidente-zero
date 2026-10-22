#!/usr/bin/env bash
# Cenário 020 — injeta a falha:
# o manifesto novo do loja-db foi copiado de homologação e veio com
# ambiente=homologacao. A CronJob de economia desliga o banco a cada 2 minutos,
# a loja perde a dependência e sai do Service (readiness), e o nginx responde 502.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja label deploy/loja-db ambiente=homologacao --overwrite >/dev/null
kubectl -n loja annotate deploy/loja-db --overwrite \
  dados.loja.local/origem="manifesto de homologação, ajustado para produção (imagem e réplicas)" >/dev/null

# A última execução da economia (como as das últimas horas)
kubectl -n loja create job --from=cronjob/economia-homologacao "economia-homologacao-$(( $(date +%s) / 60 ))" >/dev/null

# Autoverificação: banco desligado, loja sem pods prontos e 502
for _ in $(seq 1 75); do
  db=$(kubectl -n loja get deploy loja-db -o jsonpath='{.spec.replicas}' 2>/dev/null)
  prontos=$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
  code=$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)
  if [ "$db" = "0" ] && [ "${prontos:-0}" = "0" ] && [ "$code" = "502" ]; then
    exit 0
  fi
  sleep 2
done
echo "inject: a cascata não aconteceu como esperado" >&2
exit 1
