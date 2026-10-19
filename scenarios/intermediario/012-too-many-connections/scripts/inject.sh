#!/usr/bin/env bash
# Cenário 012 — injeta a falha:
# o time de BI publicou uma versão "mais rápida" do relatório: uma conexão por
# categoria, em paralelo, e uma espera pelo upload do FTP... sem fechar as conexões.
# Rodando a cada minuto, ela esgota o max_connections do banco.
set -euo pipefail
umask 022

cat > /opt/relatorios/relatorio-vendas.py <<'PY'
#!/usr/bin/env python3
"""Relatório de vendas por categoria, gerado a cada minuto para o painel do financeiro.

v2 (BI): uma conexão por categoria para paralelizar as consultas,
e espera a confirmação do upload para o FTP do financeiro.
"""
import csv
import datetime
import time

import pymysql

CATEGORIAS = ["roupas", "casa", "acessorios", "papelaria", "eletronicos"]


def conectar():
    return pymysql.connect(host="127.0.0.1", user="relatorio", password="bi-relatorio-2026", database="loja")


conexoes = {cat: conectar() for cat in CATEGORIAS}

with open("/var/lib/relatorios/vendas.csv", "w", newline="") as f:
    out = csv.writer(f)
    out.writerow(["gerado_em", "categoria", "produtos"])
    for cat, conn in conexoes.items():
        with conn.cursor() as cur:
            cur.execute("SELECT COUNT(*) FROM produtos WHERE categoria = %s", (cat,))
            out.writerow([datetime.datetime.now().isoformat(timespec="seconds"), cat, cur.fetchone()[0]])

# Aguarda a confirmação do upload para o FTP do financeiro (até 10 minutos)
time.sleep(600)
PY

# Os últimos minutos de execução do cron
for _ in $(seq 1 6); do
  setsid sudo -u relatorios /usr/bin/python3 /opt/relatorios/relatorio-vendas.py \
    >> /var/lib/relatorios/cron.log 2>&1 < /dev/null &
done
sleep 4

# Autoverificação: a loja precisa estar falhando por falta de conexões
code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' http://localhost/ || true)
if [ "$code" != "503" ]; then
  echo "inject: falha não reproduzida (HTTP $code)" >&2
  exit 1
fi
