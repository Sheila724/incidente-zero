#!/usr/bin/env bash
# Cenário 013 — prepara o ambiente SAUDÁVEL:
#   nginx (80) -> Loja Exemplo (127.0.0.1:8080) -> MariaDB, dados numa partição
#   própria (/var/lib/mysql). O binlog foi ligado para um teste de réplica.
set -euo pipefail
umask 022

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /etc/app /opt/app

cat > /etc/mysql/mariadb.conf.d/60-loja.cnf <<'EOT'
# Ajustes da Loja Exemplo
[mysqld]
bind-address            = 127.0.0.1
skip-name-resolve
innodb_buffer_pool_size = 32M
innodb_log_file_size    = 16M
performance_schema      = OFF

# Réplica de teste (CHG-340): binlog ligado para o servidor loja-db-02
server_id               = 1
log_bin                 = /var/lib/mysql/mysql-bin
binlog_format           = ROW
max_binlog_size         = 16M
# A réplica precisa de todo o histórico: não expirar os binlogs
binlog_expire_logs_seconds = 0
EOT

install -d -o mysql -g mysql -m 0750 /var/lib/mysql
chown mysql:mysql /var/lib/mysql
install -d -o mysql -g mysql -m 0755 /run/mysqld
mariadb-install-db --user=mysql --datadir=/var/lib/mysql --skip-test-db >/dev/null
systemctl enable --now mariadb

mariadb <<'SQL'
CREATE DATABASE loja;
CREATE TABLE loja.produtos (id INT PRIMARY KEY, nome VARCHAR(80), preco DECIMAL(8,2));
INSERT INTO loja.produtos VALUES (1,'Camiseta',59.90), (2,'Caneca',34.90), (3,'Boné',69.90);
CREATE TABLE loja.pedidos (id INT AUTO_INCREMENT PRIMARY KEY, produto_id INT, criado_em DATETIME DEFAULT NOW());
CREATE TABLE loja.eventos_importados (id INT AUTO_INCREMENT PRIMARY KEY, lote INT, payload LONGTEXT) ENGINE=MyISAM;
CREATE USER 'loja'@'127.0.0.1' IDENTIFIED BY 'loja-app-2026';
GRANT SELECT, INSERT ON loja.* TO 'loja'@'127.0.0.1';
SQL

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
                           database=conf["DB_NAME"], connect_timeout=3,
                           read_timeout=5, write_timeout=5)


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
            with conn, conn.cursor() as cur:
                if self.path.startswith("/checkout"):
                    cur.execute("INSERT INTO pedidos (produto_id) VALUES (1)")
                    conn.commit()
                    return self.send(200, json.dumps({"pedido": cur.lastrowid, "status": "confirmado"}))
                cur.execute("SELECT nome, preco FROM produtos ORDER BY id")
                itens = "".join(f"<li>{n} — R$ {p}</li>" for n, p in cur.fetchall())
                return self.send(200, f"<h1>Loja Exemplo</h1><ul>{itens}</ul>\n", "text/html; charset=utf-8")
        except pymysql.MySQLError as exc:
            print(f"ERRO banco ({self.path}): {exc}", flush=True)
            return self.send(503, json.dumps({"erro": "falha ao gravar no banco de dados", "detalhe": str(exc)}))

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
systemctl enable --now app
nginx -t 2>/dev/null
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 30); do
  curl -fs http://localhost/checkout | grep -q confirmado && exit 0
  sleep 0.5
done
echo "setup: a loja não respondeu" >&2
exit 1
