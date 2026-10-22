#!/usr/bin/env bash
# Cenário 012 — solução de referência (usada pelo `iz selftest`).
# Fecha as conexões antes da espera do FTP, encerra as execuções antigas
# e limita o usuário do relatório.
set -euo pipefail

python3 - <<'PY'
p = "/opt/relatorios/relatorio-vendas.py"
s = open(p).read()
s = s.replace("# Aguarda a confirmação do upload",
              "for conn in conexoes.values():\n    conn.close()\n\n# Aguarda a confirmação do upload", 1)
open(p, "w").write(s)
PY
pkill -u relatorios -f relatorio-vendas || true
mariadb -e "ALTER USER 'relatorio'@'127.0.0.1' WITH MAX_USER_CONNECTIONS 5"
