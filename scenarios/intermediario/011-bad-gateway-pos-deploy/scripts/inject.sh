#!/usr/bin/env bash
# Cenário 011 — injeta a falha:
# o deploy da versão 2.4.0 mudou a porta da aplicação para 8081,
# mas ninguém ajustou o nginx.
set -euo pipefail

cat > /etc/app/app.env <<'EOT'
# Loja Exemplo — variáveis de ambiente (gerenciado pelo deploy)
APP_VERSION=2.4.0
PORT=8081
EOT

cat > /opt/app/CHANGELOG.md <<'EOT'
# Loja Exemplo — changelog

## 2.4.0
- Página de promoções da Black Friday.
- **A aplicação passa a escutar na porta 8081.** A 8080 fica reservada para o
  agente de métricas que será instalado na próxima semana (DEP-87).
  Ajustar o proxy reverso antes do deploy (DEP-88).

## 2.3.0
- Catálogo com filtros por categoria.
EOT

systemctl restart app

# Autoverificação: o site precisa responder 502
for _ in $(seq 1 20); do
  ss -ltn | grep -q '127.0.0.1:8081' && break
  sleep 0.3
done
code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/ || true)
if [ "$code" != "502" ]; then
  echo "inject: falha não reproduzida (HTTP $code)" >&2
  exit 1
fi
