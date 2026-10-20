#!/usr/bin/env bash
# Cenário 014 — injeta a falha:
# a "padronização" do agente (role do Ansible) aplicou o arquivo da filial,
# que aponta para um proxy desativado. O agente passa a recusar o Zabbix Server.
# Enquanto o monitoramento está cego, um export de auditoria enche /var/log.
set -euo pipefail
umask 022

conf=/etc/zabbix/zabbix_agent2.conf
sed -i -e 's/^Server=.*/Server=10.20.0.50/' -e 's/^ServerActive=.*/ServerActive=10.20.0.50/' "$conf"
sed -i '1i # Gerenciado pelo Ansible (role zabbix-agent, perfil "filial"): proxy zbx-proxy-01' "$conf"
systemctl restart zabbix-agent2

# Export de auditoria gravado no lugar errado, crescendo sem ninguém ver
install -d -m 0755 /var/log/auditoria
livre_mb=$(df -Pm /var/log | awk 'NR==2 {print $4}')
total_mb=$(df -Pm /var/log | awk 'NR==2 {print $2}')
alvo=$(( livre_mb - total_mb * 6 / 100 ))
dd if=/dev/zero of=/var/log/auditoria/export-completo-2026-10.json bs=1M count="$alvo" status=none
cat > /etc/cron.d/auditoria-export <<'EOT'
# Export mensal de auditoria (time de compliance)
0 2 * * * root /usr/local/bin/auditoria-export > /var/log/auditoria/export-completo-$(date +\%Y-\%m).json
EOT

# Autoverificação: o Zabbix precisa mostrar o problema do agente
host=$(zbx-api host.get '{"filter":{"host":["loja-web-01"]},"output":["hostid"]}' | jq -r '.[0].hostid')
for _ in $(seq 1 60); do
  n=$(zbx-api problem.get "{\"hostids\":\"$host\",\"search\":{\"name\":\"Agente Zabbix\"},\"countOutput\":true}" | jq -r '.')
  [ "${n:-0}" -ge 1 ] && break
  sleep 3
done
if [ "${n:-0}" -lt 1 ]; then
  echo "inject: o problema do agente não apareceu no Zabbix" >&2
  exit 1
fi
