#!/usr/bin/env bash
# Cenário 002 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

if nginx -t >/dev/null 2>&1; then
  pass "Configuração do nginx válida (nginx -t)"
else
  fail "nginx -t ainda aponta erro na configuração"
fi

systemctl is-active --quiet nginx && pass "Serviço nginx ativo" || fail "Serviço nginx não está ativo"

# Qualquer nome (ou o IP) deve cair no site principal: ele é o default_server
body=$(curl -s -H "Host: qualquer-coisa" http://localhost/)
if echo "$body" | grep -q "<h1>Loja Exemplo</h1>"; then
  pass "Site principal respondendo como site padrão"
else
  fail "O site padrão da porta 80 não é a Loja Exemplo"
fi

body=$(curl -s -H "Host: loja.local" http://localhost/)
echo "$body" | grep -q "<h1>Loja Exemplo</h1>" && pass "loja.local respondendo" || fail "loja.local não responde a Loja Exemplo"

# Bônus: a landing page corrigida em vez de só removida
body=$(curl -s -H "Host: promo.loja.local" http://localhost/)
if echo "$body" | grep -q "Black Friday"; then
  pass "Bônus: landing page da Black Friday no ar em promo.loja.local"
fi

exit $status
