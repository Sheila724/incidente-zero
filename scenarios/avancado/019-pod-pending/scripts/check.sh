#!/usr/bin/env bash
# Cenário 019 — valida a correção. Exit 0 = resolvido.
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/checkout/)
[ "$code" = "200" ] && pass "Checkout respondendo HTTP 200" || fail "Checkout respondendo HTTP ${code:-sem resposta}"

prontos=$(kubectl -n loja get deploy loja-checkout -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
if [ "${prontos:-0}" -ge 2 ] 2>/dev/null; then
  pass "Deployment loja-checkout com ${prontos} réplicas prontas"
else
  fail "Deployment loja-checkout com ${prontos:-0} de 2 réplicas prontas"
fi

pendentes=$(kubectl -n loja get pods -l app=loja-checkout --field-selector=status.phase=Pending -o name 2>/dev/null | wc -l)
if [ "$pendentes" -eq 0 ]; then
  pass "Nenhum pod do checkout em Pending"
else
  fail "${pendentes} pod(s) do checkout em Pending"
fi

# A reserva de CPU continua existindo e é razoável (até 1 núcleo)
cpu=$(kubectl -n loja get deploy loja-checkout -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}' 2>/dev/null)
milli=""
case "$cpu" in
  "") ;;
  *m) milli=${cpu%m} ;;
  *.*) milli=$(awk -v c="$cpu" 'BEGIN { printf "%d", c * 1000 }') ;;
  *) milli=$(( cpu * 1000 )) ;;
esac
if [ -z "$cpu" ]; then
  contorno "O checkout ficou sem reserva de CPU (requests): ele sobe, mas disputa CPU com a loja no pico"
elif [ "$milli" -gt 0 ] && [ "$milli" -le 1000 ]; then
  pass "Reserva de CPU do checkout em ${cpu}"
else
  fail "Reserva de CPU do checkout em ${cpu}: confira a unidade (1 = um núcleo, 500m = meio núcleo)"
fi

read -r web_prontos web_desejadas <<< "$(kubectl -n loja get deploy loja-web -o jsonpath='{.status.readyReplicas} {.spec.replicas}' 2>/dev/null)"
if [ -z "$web_desejadas" ]; then web_desejadas=$web_prontos; web_prontos=0; fi
home=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
if [ "$home" = "200" ] && [ "${web_prontos:-0}" -ge 2 ] 2>/dev/null; then
  pass "Loja principal no ar (loja-web com ${web_prontos} réplicas prontas)"
else
  contorno "A loja principal foi prejudicada (loja-web com ${web_prontos:-0} de ${web_desejadas:-?} réplicas prontas): não se abre espaço tirando outro serviço do ar"
fi

exit $status
