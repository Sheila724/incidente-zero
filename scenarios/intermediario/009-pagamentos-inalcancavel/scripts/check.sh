#!/usr/bin/env bash
# Cenário 009 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/checkout)
[ "$code" = "200" ] && pass "Checkout respondendo HTTP 200" || fail "Checkout respondendo HTTP ${code:-sem resposta}"

resolvido=$(getent ahostsv4 api.pagamentos.interno | awk 'NR==1 {print $1}')
if [ "$resolvido" = "10.20.0.15" ]; then
  pass "O servidor resolve api.pagamentos.interno para o IP novo (10.20.0.15)"
else
  fail "O servidor resolve api.pagamentos.interno para ${resolvido:-nada}, não para 10.20.0.15"
fi

if grep -Eq '^[^#]*api\.pagamentos\.interno' /etc/hosts; then
  contorno "O /etc/hosts ainda fixa o IP da API: na próxima migração o problema volta. O nome deve vir do DNS"
else
  pass "Nenhum IP fixado no /etc/hosts: o nome vem do DNS interno"
fi

ip=$(curl -s http://127.0.0.1:8080/health | python3 -c 'import sys, json; print(json.load(sys.stdin)["pagamentos_ip"])' 2>/dev/null)
if [ "$ip" = "10.20.0.15" ]; then
  pass "A aplicação em execução usa o IP novo"
else
  fail "A aplicação em execução ainda usa ${ip:-um IP desconhecido}: ela guardou o endereço quando iniciou"
fi

systemctl is-active --quiet dns-interno && pass "DNS interno ativo" || fail "O serviço dns-interno não está ativo"

if grep -q 'PAGAMENTOS_URL=http://api.pagamentos.interno' /etc/app/app.conf; then
  pass "A aplicação continua usando o nome, não um IP fixo"
else
  fail "A configuração da aplicação não usa mais o nome api.pagamentos.interno"
fi

exit $status
