#!/usr/bin/env bash
# Cenário 004 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }

# 1. O gerador de relatórios precisa continuar ligado
if systemctl is-active --quiet report-worker; then
  pass "Gerador de relatórios ativo"
else
  fail "O report-worker está parado: o financeiro precisa dos relatórios"
fi
if systemctl is-enabled --quiet report-worker 2>/dev/null; then
  pass "Gerador de relatórios habilitado no boot"
else
  fail "O report-worker foi desabilitado e não sobe no próximo boot"
fi

# 2. Causa raiz: intervalo configurado
v=$(grep -E '^INTERVAL_SECONDS=' /etc/report-worker/worker.conf 2>/dev/null | tail -1 | cut -d= -f2 | tr -dc 0-9)
if [ -n "$v" ] && [ "$v" -ge 60 ]; then
  pass "INTERVAL_SECONDS=$v (pelo menos 60 s entre relatórios)"
else
  fail "INTERVAL_SECONDS ainda está em ${v:-?}: o worker vai rodar sem pausa"
fi

# 3. CPU do worker agora (ticks de CPU em 2 s; 100 ticks = 1 núcleo inteiro por 1 s)
pid=$(systemctl show -p MainPID --value report-worker 2>/dev/null)
if [ -n "$pid" ] && [ "$pid" -gt 0 ] && [ -r "/proc/$pid/stat" ]; then
  a=$(awk '{print $14+$15}' "/proc/$pid/stat"); sleep 2; b=$(awk '{print $14+$15}' "/proc/$pid/stat")
  if [ $((b - a)) -lt 10 ]; then
    pass "Worker em repouso entre relatórios (CPU baixa)"
  else
    fail "O worker em execução ainda está consumindo CPU (reiniciou o serviço depois de corrigir?)"
  fi
fi

# 4. Relatórios continuam sendo gerados
if [ -n "$(find /var/lib/report-worker/report-latest.json -mmin -10 2>/dev/null)" ]; then
  pass "Último relatório gerado há menos de 10 minutos"
else
  fail "Nenhum relatório recente em /var/lib/report-worker"
fi

# 5. Loja respondendo rápido
t=$(curl -s -o /dev/null -m 5 -w '%{http_code} %{time_total}' http://localhost/)
code=${t%% *}; secs=${t#* }
if [ "$code" = "200" ] && awk -v s="$secs" 'BEGIN{exit !(s < 1.0)}'; then
  pass "Loja respondendo 200 em ${secs}s"
else
  fail "Loja respondendo ${code:-000} em ${secs:-?}s"
fi

# Bônus: limite de CPU no serviço
q=$(systemctl show -p CPUQuotaPerSecUSec --value report-worker 2>/dev/null)
if [ -n "$q" ] && [ "$q" != "infinity" ]; then
  pass "Bônus: report-worker com limite de CPU (CPUQuota)"
fi

exit $status
