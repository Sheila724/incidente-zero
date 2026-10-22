#!/usr/bin/env bash
# Cenário 018 — valida a correção. Exit 0 = resolvido.
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-sem resposta}"

prontos=$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
if [ "${prontos:-0}" -ge 2 ] 2>/dev/null; then
  pass "Deployment loja-web com ${prontos} réplicas prontas"
else
  fail "Deployment loja-web com ${prontos:-0} de 2 réplicas prontas"
fi

eps=$(kubectl -n loja get endpointslices -l kubernetes.io/service-name=loja-web \
  -o jsonpath='{range .items[*].endpoints[?(@.conditions.ready==true)]}{.addresses[0]}{"\n"}{end}' 2>/dev/null | grep -c .)
if [ "$eps" -ge 2 ]; then
  pass "Service loja-web com ${eps} endpoints prontos"
else
  fail "Service loja-web com ${eps} endpoints prontos"
fi

probe=$(kubectl -n loja get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].readinessProbe}' 2>/dev/null)
caminho=$(kubectl -n loja get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].readinessProbe.httpGet.path}' 2>/dev/null)
if [ -z "$probe" ]; then
  fail "A readiness probe foi removida: sem ela, a loja recebe clientes antes de estar pronta"
elif [ -n "$caminho" ]; then
  pod=$(kubectl -n loja get pods -l app=loja-web --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
  if [ -n "$pod" ] && kubectl -n loja exec "$pod" -- wget -q -O /dev/null "http://localhost:8080${caminho}" 2>/dev/null; then
    pass "Readiness probe verificando um caminho que a aplicação atende (${caminho})"
  else
    fail "A readiness probe chama ${caminho}, que a aplicação não atende"
  fi
else
  pass "Readiness probe configurada"
fi

exit $status
