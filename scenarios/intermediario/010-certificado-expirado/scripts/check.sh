#!/usr/bin/env bash
# Cenário 010 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

servido=$(mktemp)
trap 'rm -f "$servido"' EXIT
openssl s_client -connect 127.0.0.1:443 -servername loja.local </dev/null 2>/dev/null \
  | openssl x509 > "$servido" 2>/dev/null

if [ ! -s "$servido" ]; then
  fail "O nginx não está servindo HTTPS na porta 443"
elif openssl x509 -checkend $((30 * 86400)) -noout -in "$servido" >/dev/null; then
  pass "O certificado servido vale até $(openssl x509 -enddate -noout -in "$servido" | cut -d= -f2)"
else
  fail "O certificado que o nginx está servindo vence em menos de 30 dias (ou já venceu)"
fi

if curl -fs -o /dev/null -m 10 --resolve loja.local:443:127.0.0.1 https://loja.local/; then
  pass "https://loja.local responde com um certificado válido e confiável"
else
  fail "https://loja.local ainda falha na verificação do certificado"
fi

if [ -s "$servido" ] && [ "$(openssl x509 -noout -fingerprint -sha256 -in "$servido")" = \
     "$(openssl x509 -noout -fingerprint -sha256 -in /etc/ssl/loja/loja.crt 2>/dev/null)" ]; then
  pass "O nginx está servindo o certificado que está no disco"
else
  fail "O certificado no disco é diferente do que o nginx está servindo"
fi

if systemctl is-enabled --quiet renova-certificado.timer && systemctl is-active --quiet renova-certificado.timer; then
  pass "Renovação automática habilitada (renova-certificado.timer)"
else
  fail "O timer renova-certificado não está habilitado e ativo: o certificado vai vencer de novo"
fi

if grep -Eq '^[^#]*(systemctl +(reload|restart) +nginx|nginx +-s +reload)' /usr/local/sbin/renova-certificado \
   || systemctl cat renova-certificado.service 2>/dev/null | grep -Eq '^ExecStartPost=.*nginx'; then
  pass "Depois de renovar, o nginx é recarregado"
else
  fail "A renovação grava o certificado novo, mas não recarrega o nginx: ele continuaria servindo o antigo"
fi

if [ "$(stat -c %a /etc/ssl/loja/loja.key)" = "600" ]; then
  pass "Chave privada protegida (600)"
else
  fail "A chave privada /etc/ssl/loja/loja.key deveria ter permissão 600"
fi

exit $status
