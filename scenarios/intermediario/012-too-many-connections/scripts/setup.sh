#!/usr/bin/env bash
# Cenário 012 — prepara o ambiente SAUDÁVEL:
#   nginx (80) -> Loja Exemplo (127.0.0.1:8080) -> MariaDB (banco "loja")
#   Um job de relatório de vendas roda pelo cron a cada minuto.
set -euo pipefail
umask 022

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
id relatorios &>/dev/null || useradd --system --create-home --home-dir /var/lib/relatorios --shell /usr/sbin/nologin relatorios
install -d -m 0755 /etc/app /opt/app /opt/relatorios

# ---------- MariaDB ----------
cat > /etc/mysql/mariadb.conf.d/60-loja.cnf <<'EOT'
# Ajustes da Loja Exemplo (servidor pequeno)
[mysqld]
bind-address            = 127.0.0.1
skip-name-resolve
innodb_buffer_pool_size = 32M
innodb_log_file_size    = 16M
performance_schema      = OFF
# Reduzido no tuning de memória (CHG-311): cada conexão consome RAM
max_connections         = 30
EOT

install -d -o mysql -g mysql -m 0750 /var/lib/mysql
install -d -o mysql -g mysql -m 0755 /run/mysqld
mariadb-install-db --user=mysql --datadir=/var/lib/mysql --skip-test-db >/dev/null
systemctl enable --now mariadb

mariadb <<'SQL'
CREATE DATABASE loja;
CREATE TABLE loja.produtos (id INT PRIMARY KEY, nome VARCHAR(80), preco DECIMAL(8,2), categoria VARCHAR(30));
INSERT INTO loja.produtos VALUES
  (1,'Camiseta',59.90,'roupas'), (2,'Moletom',149.90,'roupas'),
  (3,'Caneca',34.90,'casa'), (4,'Almofada',49.90,'casa'),
  (5,'Adesivo',9.90,'acessorios'), (6,'Boné',69.90,'acessorios');
CREATE TABLE loja.pedidos (id INT AUTO_INCREMENT PRIMARY KEY, produto_id INT, criado_em DATETIME DEFAULT NOW());
CREATE USER 'loja'@'127.0.0.1' IDENTIFIED BY 'loja-app-2026';
GRANT SELECT, INSERT ON loja.* TO 'loja'@'127.0.0.1';
CREATE USER 'relatorio'@'127.0.0.1' IDENTIFIED BY 'bi-relatorio-2026';
GRANT SELECT ON loja.* TO 'relatorio'@'127.0.0.1';
SQL

# ---------- Loja Exemplo ----------
cat > /etc/app/app.conf <<'EOT'
DB_HOST=127.0.0.1
DB_USER=loja
DB_PASSWORD=loja-app-2026
DB_NAME=loja
EOT
chown root:app /etc/app/app.conf
chmod 0640 /etc/app/app.conf

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: catálogo e checkout em cima do MariaDB."""
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import pymysql

conf = dict(l.strip().split("=", 1) for l in open("/etc/app/app.conf") if "=" in l)


def db():
    return pymysql.connect(host=conf["DB_HOST"], user=conf["DB_USER"], password=conf["DB_PASSWORD"],
                           database=conf["DB_NAME"], connect_timeout=3)


class H(BaseHTTPRequestHandler):
    def send(self, code, body, ctype="application/json"):
        data = body.encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        try:
            conn = db()
        except pymysql.MySQLError as exc:
            print(f"ERRO banco: {exc}", flush=True)
            return self.send(503, json.dumps({"erro": "banco de dados indisponível", "detalhe": str(exc)}))
        try:
            with conn.cursor() as cur:
                if self.path.startswith("/checkout"):
                    cur.execute("INSERT INTO pedidos (produto_id) VALUES (1)")
                    conn.commit()
                    return self.send(200, json.dumps({"pedido": cur.lastrowid, "status": "confirmado"}))
                cur.execute("SELECT nome, preco FROM produtos ORDER BY id")
                itens = "".join(f"<li>{n} — R$ {p}</li>" for n, p in cur.fetchall())
                return self.send(200, f"<h1>Loja Exemplo</h1><ul>{itens}</ul>\n", "text/html; charset=utf-8")
        finally:
            conn.close()

    def log_message(self, *a):
        pass


ThreadingHTTPServer(("127.0.0.1", 8080), H).serve_forever()
PY

cat > /etc/systemd/system/app.service <<'EOT'
[Unit]
Description=Loja Exemplo
After=network.target mariadb.service

[Service]
User=app
ExecStart=/usr/bin/python3 /opt/app/app.py
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOT

# ---------- Relatório de vendas (versão estável) ----------
cat > /opt/relatorios/relatorio-vendas.py <<'PY'
#!/usr/bin/env python3
"""Relatório de vendas por categoria, gerado a cada minuto para o painel do financeiro."""
import csv
import datetime

import pymysql

CATEGORIAS = ["roupas", "casa", "acessorios", "papelaria", "eletronicos"]

conn = pymysql.connect(host="127.0.0.1", user="relatorio", password="bi-relatorio-2026", database="loja")
with conn, open("/var/lib/relatorios/vendas.csv", "w", newline="") as f:
    out = csv.writer(f)
    out.writerow(["gerado_em", "categoria", "produtos"])
    with conn.cursor() as cur:
        for cat in CATEGORIAS:
            cur.execute("SELECT COUNT(*) FROM produtos WHERE categoria = %s", (cat,))
            out.writerow([datetime.datetime.now().isoformat(timespec="seconds"), cat, cur.fetchone()[0]])
PY
chmod 0755 /opt/relatorios/relatorio-vendas.py

cat > /etc/cron.d/relatorio-vendas <<'EOT'
# Relatório de vendas para o painel do financeiro (time de BI)
* * * * * relatorios /usr/bin/python3 /opt/relatorios/relatorio-vendas.py >> /var/lib/relatorios/cron.log 2>&1
EOT
chmod 0644 /etc/cron.d/relatorio-vendas

cat > /etc/nginx/sites-available/loja <<'EOT'
server {
    listen 80 default_server;
    server_name _;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }
}
EOT
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/loja /etc/nginx/sites-enabled/loja

systemctl daemon-reload
systemctl enable --now app cron
nginx -t 2>/dev/null
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 30); do
  curl -fs http://localhost/checkout | grep -q confirmado && exit 0
  sleep 0.5
done
echo "setup: a loja não respondeu" >&2
exit 1
