#!/usr/bin/env bash
# Cenário 012 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

systemctl is-active --quiet mariadb && pass "MariaDB ativo" || fail "O MariaDB não está ativo"

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/checkout)
[ "$code" = "200" ] && pass "Checkout respondendo HTTP 200" || fail "Checkout respondendo HTTP ${code:-sem resposta}"

if grep -Eq '^[^#]*relatorio-vendas\.py' /etc/cron.d/relatorio-vendas 2>/dev/null; then
  pass "O relatório do financeiro continua agendado"
else
  fail "O relatório foi desligado: o financeiro depende dele. Corrija a causa em vez de desativar o job"
fi

conexoes_relatorio() {
  mariadb -N -e "SELECT COUNT(*) FROM information_schema.processlist WHERE user = 'relatorio'" 2>/dev/null
}

# Uma execução do relatório: depois de gerar o arquivo, não pode segurar conexões
sudo -u relatorios /usr/bin/python3 /opt/relatorios/relatorio-vendas.py >/dev/null 2>&1 < /dev/null &
um=$!
sleep 3
presas=$(conexoes_relatorio)
kill "$um" 2>/dev/null
wait "$um" 2>/dev/null
if [ "${presas:-1}" = "0" ]; then
  pass "O relatório fecha as conexões depois de usar"
else
  fail "Há ${presas:-?} conexões do relatório presas: o script ainda não fecha as conexões, ou execuções antigas continuam rodando (ps -ef | grep relatorio)"
fi

# Simula os próximos 15 minutos do cron, todos de uma vez
rm -f /var/lib/relatorios/vendas.csv
pids=()
for _ in $(seq 1 15); do
  sudo -u relatorios /usr/bin/python3 /opt/relatorios/relatorio-vendas.py >/dev/null 2>&1 < /dev/null &
  pids+=("$!")
done
sleep 5
code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/checkout)
# Encerra só as execuções simuladas por esta verificação
kill "${pids[@]}" 2>/dev/null
wait "${pids[@]}" 2>/dev/null

if [ "$code" = "200" ]; then
  pass "Com o relatório rodando várias vezes, a loja continua conectando no banco"
else
  fail "Depois de algumas execuções do relatório, a loja volta a falhar (HTTP ${code:-sem resposta}): o banco enche de novo"
fi

if [ -s /var/lib/relatorios/vendas.csv ]; then
  pass "O relatório de vendas continua sendo gerado"
else
  fail "O relatório não gerou o vendas.csv"
fi

limite=$(mariadb -N -e "SELECT max_user_connections FROM mysql.user WHERE user = 'relatorio' LIMIT 1" 2>/dev/null)
if [ "${limite:-0}" -gt 0 ] 2>/dev/null; then
  pass "Bônus: o usuário relatorio tem limite próprio de conexões (${limite})"
fi

exit $status
