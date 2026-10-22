#!/usr/bin/env bash
# Cenário 013 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

systemctl is-active --quiet mariadb && pass "MariaDB ativo" || fail "O MariaDB não está ativo"

uso=$(df --output=pcent /var/lib/mysql | tail -1 | tr -dc '0-9')
if [ "${uso:-100}" -lt 80 ]; then
  pass "Partição do banco com espaço (${uso}% em uso)"
else
  fail "A partição /var/lib/mysql continua com ${uso}% de uso"
fi

code=$(curl -s -o /dev/null -m 15 -w '%{http_code}' http://localhost/checkout)
[ "$code" = "200" ] && pass "Checkout gravando pedidos (HTTP 200)" || fail "Checkout respondendo HTTP ${code:-sem resposta}"

# O índice do binlog precisa bater com os arquivos (rm na mão quebra isso)
indice=/var/lib/mysql/mysql-bin.index
if [ -f "$indice" ]; then
  faltando=0
  while read -r arq; do
    [ -z "$arq" ] && continue
    case "$arq" in /*) caminho=$arq ;; *) caminho=/var/lib/mysql/${arq#./} ;; esac
    [ -f "$caminho" ] || faltando=$((faltando + 1))
  done < "$indice"
  if [ "$faltando" -eq 0 ]; then
    pass "Índice do binlog consistente com os arquivos no disco"
  else
    contorno "O índice do binlog lista $faltando arquivo(s) que não existem mais: binlogs apagados com rm em vez de PURGE BINARY LOGS"
  fi
else
  pass "Binlog desligado: não há índice para conferir"
fi

# Prevenção: binlog com expiração, ou desligado, na configuração (sobrevive a restart)
opcoes=$(my_print_defaults --mysqld 2>/dev/null)
log_bin_cfg=$(echo "$opcoes" | grep -Ec '^--log[_-]bin(=|$)')
skip_cfg=$(echo "$opcoes" | grep -Ec '^--(skip[_-]log[_-]bin|disable[_-]log[_-]bin)$')
expira_cfg=$(echo "$opcoes" | grep -E '^--(binlog[_-]expire[_-]logs[_-]seconds|expire[_-]logs[_-]days)=' | tail -1 | cut -d= -f2)
if [ "$log_bin_cfg" -eq 0 ] || [ "$skip_cfg" -gt 0 ]; then
  pass "Configuração: binlog desligado (a réplica de teste não existe mais)"
elif [ -n "$expira_cfg" ] && awk -v v="$expira_cfg" 'BEGIN { exit !(v + 0 > 0) }'; then
  pass "Configuração: binlog com expiração automática ($expira_cfg)"
else
  fail "A configuração mantém o binlog ligado sem expiração: o disco vai encher de novo"
fi

exp_atual=$(mariadb -N -e "SELECT @@binlog_expire_logs_seconds" 2>/dev/null)
bin_atual=$(mariadb -N -e "SELECT @@log_bin" 2>/dev/null)
if [ "$bin_atual" = "0" ] || [ "${exp_atual:-0}" -gt 0 ] 2>/dev/null; then
  pass "Servidor em execução já com a política de binlog aplicada"
else
  fail "O MariaDB em execução ainda está com binlog sem expiração (aplique com SET GLOBAL ou reinicie o serviço)"
fi

exit $status
