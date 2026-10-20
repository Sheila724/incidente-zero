#!/usr/bin/env bash
# Cenário 015 — injeta a falha:
# o deploy 2.5.0 trouxe um cache de sessões em memória, ligado sem limite.
# Cada visita guarda ~50 KB que nunca são liberados.
set -euo pipefail
umask 022

cat > /etc/app/app.env <<'EOT'
# Loja Exemplo — gerenciado pelo deploy
APP_VERSION=2.5.0
# Cache de sessões em memória (novo na 2.5.0)
SESSION_CACHE_MAX=0
EOT

cat > /opt/app/CHANGELOG.md <<'EOT'
# Loja Exemplo — changelog

## 2.5.0
- Cache de sessões em memória: o carrinho abre até 10x mais rápido.
- Novo parâmetro `SESSION_CACHE_MAX` em /etc/app/app.env: quantas sessões
  ficam em memória (as mais antigas saem primeiro). `0` = sem limite.
  Valor recomendado para produção: 2000.

## 2.4.0
- Página de promoções.
EOT
systemctl restart app

# As últimas horas de tráfego, de uma vez
for _ in $(seq 1 30); do curl -fs -o /dev/null http://localhost/ && break; sleep 0.5; done
seq 1 6000 | xargs -P 8 -I{} curl -s -o /dev/null http://127.0.0.1:8080/

# Autoverificação: o Zabbix precisa mostrar o problema de memória
for _ in $(seq 1 40); do
  n=$(zbx-api problem.get '{"search":{"name":"usando muita memória"},"countOutput":true}' | jq -r '.')
  [ "${n:-0}" -ge 1 ] && exit 0
  sleep 3
done
echo "inject: o problema de memória não apareceu no Zabbix" >&2
exit 1
