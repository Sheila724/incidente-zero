#!/usr/bin/env bash
# Cenário 020 — valida a correção. Exit 0 = resolvido.
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Loja respondendo HTTP 200" || fail "Loja respondendo HTTP ${code:-sem resposta}"

prontos=$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
if [ "${prontos:-0}" -ge 2 ] 2>/dev/null; then
  pass "loja-web com ${prontos} réplicas prontas"
else
  fail "loja-web com ${prontos:-0} de 2 réplicas prontas"
fi

db=$(kubectl -n loja get deploy loja-db -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
if [ "${db:-0}" -ge 1 ] 2>/dev/null; then
  pass "Banco loja-db no ar"
else
  fail "Banco loja-db sem réplicas prontas"
fi

amb=$(kubectl -n loja get deploy loja-db -o jsonpath='{.metadata.labels.ambiente}' 2>/dev/null)
if [ "$amb" = "homologacao" ]; then
  fail "loja-db ainda tem o rótulo ambiente=homologacao: a economia vai desligá-lo de novo em até 2 minutos"
else
  pass "loja-db fora do alcance da economia de homologação (ambiente=${amb:-sem rótulo})"
fi

if ! kubectl -n loja get cronjob economia-homologacao >/dev/null 2>&1; then
  contorno "A CronJob economia-homologacao foi apagada: a loja fica de pé, mas os ambientes de homologação voltam a gastar"
else
  suspensa=$(kubectl -n loja get cronjob economia-homologacao -o jsonpath='{.spec.suspend}' 2>/dev/null)
  if [ "$suspensa" = "true" ]; then
    contorno "A CronJob economia-homologacao está suspensa: boa mitigação, mas a causa é o rótulo do banco. Reative a automação"
  else
    pass "Automação de economia ativa"
  fi
fi

if kubectl -n loja get deploy relatorios-web -o jsonpath='{.metadata.labels.ambiente}' 2>/dev/null | grep -qx homologacao; then
  pass "relatorios-web continua marcado como homologação"
else
  contorno "relatorios-web perdeu o rótulo ambiente=homologacao: ele sairia do alcance da economia"
fi

exit $status
