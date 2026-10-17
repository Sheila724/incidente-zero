#!/usr/bin/env bash
# Cenário 005 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

body=$(curl -s -m 3 http://localhost/)
if echo "$body" | grep -q "Directory listing"; then
  fail "A porta 80 ainda está expondo a listagem de arquivos do servidor"
else
  pass "Nenhuma listagem de arquivos exposta"
fi

if systemctl is-active --quiet debug-http; then
  fail "O serviço debug-http ainda está rodando"
else
  pass "Serviço debug-http parado"
fi

state=$(systemctl is-enabled debug-http 2>/dev/null)
case "$state" in
  enabled|enabled-runtime|alias|static|indirect|generated)
    fail "O debug-http ainda está habilitado ($state): volta no próximo boot" ;;
  *) pass "O debug-http não sobe mais no boot" ;;
esac

systemctl is-active --quiet nginx && pass "Serviço nginx ativo" || fail "Serviço nginx não está ativo"
systemctl is-enabled --quiet nginx && pass "nginx habilitado no boot" || fail "nginx não está habilitado no boot"

if echo "$body" | grep -q "<h1>Loja Exemplo</h1>"; then
  pass "Loja Exemplo respondendo na porta 80"
else
  fail "A porta 80 não está servindo a Loja Exemplo"
fi

# Bônus: o serviço foi removido ou mascarado, e não só desabilitado
if [ "$state" = "masked" ] || [ ! -e /etc/systemd/system/debug-http.service ]; then
  pass "Bônus: debug-http removido ou mascarado de vez"
fi

exit $status
