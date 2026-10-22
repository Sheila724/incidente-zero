#!/usr/bin/env bash
# Cenário 003 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

site_ok() {
  for p in / /produtos/ /img/logo.svg; do
    [ "$(curl -s -o /dev/null -w '%{http_code}' "http://localhost$p")" = "200" ] || return 1
  done
}

systemctl is-active --quiet nginx && pass "Serviço nginx ativo" || fail "Serviço nginx não está ativo"

for p in / /produtos/ /img/logo.svg; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "http://localhost$p")
  [ "$code" = "200" ] && pass "$p respondendo 200" || fail "$p respondendo $code"
done

# Segurança: corrigir com 777 não vale
if [ -n "$(find /var/www/loja -perm -o+w 2>/dev/null | head -1)" ]; then
  contorno "Há arquivos ou pastas graváveis por qualquer usuário (o+w): isso abre uma falha de segurança"
else
  pass "Nenhum arquivo do site gravável por qualquer usuário"
fi

# Recorrência: o agendamento não pode quebrar o site de novo
if systemctl is-active --quiet hardening-web.timer; then
  systemctl start hardening-web.service 2>/dev/null
  if site_ok; then
    pass "O hardening agendado roda e o site continua no ar"
    if [ -z "$(find /var/www/loja -perm /o+r 2>/dev/null | head -1)" ]; then
      pass "Bônus: site no ar com permissões mínimas (sem leitura para 'outros')"
    fi
  else
    fail "O timer hardening-web ainda está ativo e derruba o site de novo (acabei de rodá-lo)"
  fi
else
  pass "Agendamento hardening-web desativado (lembre-se: a auditoria SEC-142 continua pendente)"
fi

exit $status
