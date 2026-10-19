#!/usr/bin/env bash
# Cenário 011 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

if nginx -t >/dev/null 2>&1; then
  pass "Configuração do nginx válida"
else
  fail "nginx -t aponta erro na configuração"
fi
systemctl is-active --quiet nginx && pass "Serviço nginx ativo" || fail "Serviço nginx não está ativo"

code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/)
[ "$code" = "200" ] && pass "Site respondendo HTTP 200" || fail "Site respondendo HTTP ${code:-sem resposta}"

api=$(curl -s -m 10 http://localhost/api/produtos)
if echo "$api" | grep -q '"versao": "2.4.0"'; then
  pass "API de produtos respondendo com a versão 2.4.0"
else
  fail "A API (/api/produtos) não responde com a versão 2.4.0"
fi

if grep -q '^PORT=8081' /etc/app/app.env && ss -ltn | grep -q '127.0.0.1:8081 '; then
  pass "Aplicação na porta 8081, como definido no deploy 2.4.0"
else
  fail "A aplicação deveria escutar na 8081: a porta 8080 será usada pelo agente de métricas (DEP-87)"
fi

# Bônus: um único lugar define o backend
if ! grep -REq '^[^#]*proxy_pass +http://127\.0\.0\.1:' /etc/nginx/sites-enabled/ /etc/nginx/conf.d/ 2>/dev/null; then
  pass "Bônus: todas as rotas usam o upstream loja_app (o endereço fica num lugar só)"
fi

exit $status
