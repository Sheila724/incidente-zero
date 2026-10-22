#!/usr/bin/env bash
# Cenário 017 — valida a correção. Exit 0 = resolvido.
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-sem resposta}"

img=$(kubectl -n loja get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null)
case "$img" in
  registry.loja.local/loja-web:2.6) pass "Deployment usando a imagem publicada: $img" ;;
  *:v2.6) fail "Deployment ainda pede uma imagem que não existe: $img" ;;
  *) fail "A versão 2.6 não está no ar (imagem: ${img:-?}). Desfazer o deploy não atende o marketing" ;;
esac

read -r desejadas atualizadas prontas total <<< "$(kubectl -n loja get deploy loja-web \
  -o jsonpath='{.spec.replicas} {.status.updatedReplicas} {.status.readyReplicas} {.status.replicas}' 2>/dev/null)"
if [ "${desejadas:-0}" -ge 2 ] && [ "${atualizadas:-0}" = "$desejadas" ] && [ "${prontas:-0}" = "$desejadas" ] && [ "${total:-0}" = "$desejadas" ]; then
  pass "Rollout concluído: ${prontas} de ${desejadas} réplicas prontas na versão nova"
else
  fail "Rollout incompleto: ${prontas:-0} prontas, ${atualizadas:-0} atualizadas, ${total:-0} pods para ${desejadas:-?} réplicas"
fi

ruins=$(kubectl -n loja get pods -l app=loja-web \
  -o jsonpath='{range .items[*]}{.metadata.name} {.status.containerStatuses[0].state.waiting.reason}{"\n"}{end}' 2>/dev/null \
  | awk '$2 != "" {print $1" ("$2")"}')
if [ -z "$ruins" ]; then
  pass "Nenhum pod esperando imagem"
else
  while IFS= read -r p; do fail "Pod com problema: $p"; done <<< "$ruins"
fi

if curl -s -m 5 http://localhost/health | grep -q 'versao=2.6'; then
  pass "Os clientes estão vendo a versão 2.6"
else
  fail "Os clientes ainda não estão vendo a versão 2.6"
fi

exit $status
