#!/usr/bin/env bash
# Cenário 016 — valida a correção. Exit 0 = resolvido.
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-sem resposta}"

# A aplicação recebe DATABASE_URL? (pela ConfigMap ou direto no Deployment)
cm=$(kubectl -n loja get configmap loja-config -o jsonpath='{.data.DATABASE_URL}' 2>/dev/null)
usa_cm=$(kubectl -n loja get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].envFrom[*].configMapRef.name}' 2>/dev/null)
direto=$(kubectl -n loja get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="DATABASE_URL")].name}' 2>/dev/null)
config_ok=0
if { [ -n "$cm" ] && grep -qw loja-config <<< "$usa_cm"; } || [ -n "$direto" ]; then
  config_ok=1
  pass "A configuração entrega DATABASE_URL para a aplicação"
else
  fail "A configuração não entrega DATABASE_URL para a aplicação"
fi

prontos=$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
if [ "${prontos:-0}" -ge 2 ] 2>/dev/null; then
  pass "Deployment loja-web com ${prontos} réplicas prontas"
else
  fail "Deployment loja-web com ${prontos:-0} de 2 réplicas prontas"
fi

ruins=$(kubectl -n loja get pods -l app=loja-web --field-selector=status.phase!=Succeeded \
  -o jsonpath='{range .items[*]}{.metadata.name} {.status.containerStatuses[0].state.waiting.reason}{"\n"}{end}' 2>/dev/null \
  | awk '$2 != "" {print $1" ("$2")"}')
if [ -z "$ruins" ]; then
  pass "Nenhum pod da loja-web em CrashLoopBackOff ou erro"
else
  while IFS= read -r p; do fail "Pod com problema: $p"; done <<< "$ruins"
  if [ "$config_ok" = 1 ]; then
    fail "A configuração já está certa, mas os pods em execução foram criados antes dela: variáveis de ambiente só são lidas quando o container inicia"
  fi
fi

health=$(curl -s -m 5 http://localhost/health)
if grep -q 'db=mysql://' <<< "$health"; then
  pass "A aplicação subiu com a conexão do banco configurada"
else
  fail "A aplicação não está com a conexão do banco configurada"
fi

exit $status
