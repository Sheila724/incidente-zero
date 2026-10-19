#!/usr/bin/env bash
# Cenário 009 — injeta a falha:
# a API de pagamentos migrou de pag-api-01 (10.20.0.9) para pag-api-02 (10.20.0.15).
# O DNS foi atualizado, mas o /etc/hosts ainda fixa o IP antigo e a aplicação
# guardou o IP antigo na memória quando iniciou.
set -euo pipefail

# Migração: o servidor antigo foi desligado e o novo assumiu
systemctl stop pagamentos-api
ip addr del 10.20.0.9/32 dev lo 2>/dev/null || true
sed -i 's/^10\.20\.0\.9 .*/10.20.0.15   pag-api-02/' /etc/rede-interna/enderecos
sed -i 's/^BIND=.*/BIND=10.20.0.15/' /etc/pagamentos.conf
/usr/local/sbin/rede-interna
systemctl start pagamentos-api

# DNS atualizado corretamente
sed -i 's/^\(api\.pagamentos\.interno *A *\).*/\110.20.0.15/' /etc/dns-interno/interno.zone
systemctl reload dns-interno

# Autoverificação: o checkout precisa estar falhando, com o DNS já correto
for _ in $(seq 1 20); do
  dig +short @10.20.0.2 api.pagamentos.interno | grep -qx 10.20.0.15 && break
  sleep 0.3
done
code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/checkout || true)
if [ "$code" != "502" ]; then
  echo "inject: falha não reproduzida (checkout HTTP $code)" >&2
  exit 1
fi
