#!/usr/bin/env bash
# Cenário 006 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

systemctl is-active --quiet cron && pass "Serviço cron ativo" || fail "O serviço cron não está ativo"

# Onde está agendado o backup (cron.d com nome válido ou crontab do root)
valid=""
for f in /etc/cron.d/*; do
  [ -f "$f" ] || continue
  grep -q 'backup-loja.sh' "$f" || continue
  if basename "$f" | grep -Eq '^[A-Za-z0-9_-]+$'; then valid="$f"; fi
done
if [ -z "$valid" ] && crontab -l -u root 2>/dev/null | grep -q 'backup-loja.sh'; then
  valid="crontab do root"
fi

if [ -n "$valid" ]; then
  pass "Agendamento do backup num local que o cron lê ($valid)"
else
  fail "Nenhum agendamento do backup que o cron leia (confira o nome do arquivo no /etc/cron.d)"
fi

# Linha sem % sem escape
if [ -n "$valid" ]; then
  if [ "$valid" = "crontab do root" ]; then
    line=$(crontab -l -u root | grep 'backup-loja.sh' | grep -v '^#' | head -1)
  else
    line=$(grep 'backup-loja.sh' "$valid" | grep -v '^#' | head -1)
  fi
  if echo "$line" | grep -Eq '(^|[^\\])%'; then
    fail "A linha do agendamento ainda tem % sem escape: o comando é cortado"
  else
    pass "Linha do agendamento sem % sem escape"
  fi
fi

# Backup recente e válido
latest=$(ls -t /var/backups/loja/loja-*.tar.gz 2>/dev/null | head -1)
if [ -n "$latest" ] && [ -n "$(find "$latest" -mmin -5)" ]; then
  if tar -tzf "$latest" 2>/dev/null | grep -q 'data/VERSION'; then
    pass "Backup recente e válido: $(basename "$latest")"
  else
    fail "O backup mais recente está corrompido ou incompleto"
  fi
else
  fail "Nenhum backup nos últimos 5 minutos (se acabou de corrigir, aguarde o próximo ciclo de 2 min)"
fi

# O backup veio do cron, e não só da execução manual
if journalctl -u cron --since "-6 min" --no-pager 2>/dev/null | grep -q 'CMD (/usr/local/bin/backup-loja.sh'; then
  pass "O cron executou o backup nos últimos minutos"
else
  fail "O journal do cron não mostra o backup executando: ele ainda não roda sozinho"
fi

# Bônus: a saída da execução está sendo guardada
if grep -qs 'backup ok' /var/log/backup-loja*.log; then
  pass "Bônus: log da execução guardado em /var/log"
fi

exit $status
