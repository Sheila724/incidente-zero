#!/usr/bin/env bash
# Cenário 009 — prepara o ambiente SAUDÁVEL:
#   nginx (80) -> Loja Exemplo (127.0.0.1:8080) -> API de pagamentos (api.pagamentos.interno:9000)
#   DNS interno (10.20.0.2) responde a zona "interno".
#   A API de pagamentos ainda está no servidor antigo, 10.20.0.9.
set -euo pipefail
umask 022

id app &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin app
install -d -m 0755 /etc/app /opt/app /opt/pagamentos /opt/dns-interno /etc/dns-interno /etc/rede-interna

# ---------- rede interna simulada (endereços na interface lo) ----------
cat > /etc/rede-interna/enderecos <<'EOT'
# Endereços da rede interna que este servidor simula
10.20.0.2    dns-interno
10.20.0.9    pag-api-01
EOT

cat > /usr/local/sbin/rede-interna <<'EOT'
#!/usr/bin/env bash
# Sobe os endereços listados em /etc/rede-interna/enderecos
grep -Ev '^\s*(#|$)' /etc/rede-interna/enderecos | while read -r ip _; do
  ip addr add "$ip/32" dev lo 2>/dev/null || true
done
EOT
chmod 0755 /usr/local/sbin/rede-interna

cat > /etc/systemd/system/rede-interna.service <<'EOT'
[Unit]
Description=Endereços da rede interna
Before=network.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/sbin/rede-interna

[Install]
WantedBy=multi-user.target
EOT

# ---------- DNS interno ----------
cat > /etc/dns-interno/interno.zone <<'EOT'
; Zona "interno" — DNS interno da empresa
; nome                         tipo  endereço
dns.interno                    A     10.20.0.2
loja.interno                   A     127.0.0.1
api.pagamentos.interno         A     10.20.0.9
EOT

cat > /opt/dns-interno/dns.py <<'PY'
#!/usr/bin/env python3
"""DNS interno mínimo: responde registros A da zona em /etc/dns-interno/interno.zone."""
import signal
import socket
import struct
import sys

ZONE = "/etc/dns-interno/interno.zone"
records = {}


def load(*_):
    global records
    new = {}
    with open(ZONE) as f:
        for line in f:
            line = line.split(";", 1)[0].strip()
            parts = line.split()
            if len(parts) == 3 and parts[1].upper() == "A":
                new.setdefault(parts[0].lower().rstrip("."), []).append(parts[2])
    records = new
    print(f"zona carregada: {len(records)} nomes", flush=True)


def parse_name(data, off):
    labels = []
    while data[off]:
        n = data[off]
        labels.append(data[off + 1:off + 1 + n].decode())
        off += n + 1
    return ".".join(labels).lower(), off + 1


def answer(query):
    qid, _flags, qd = struct.unpack("!HHH", query[:6])
    if qd != 1:
        return None
    name, off = parse_name(query, 12)
    qtype, _qclass = struct.unpack("!HH", query[off:off + 4])
    question = query[12:off + 4]
    ips = records.get(name)
    rcode = 0 if ips else 3                     # 3 = NXDOMAIN
    answers = b""
    if ips and qtype == 1:                       # registro A
        for ip in ips:
            answers += struct.pack("!HHHIH", 0xC00C, 1, 1, 60, 4) + socket.inet_aton(ip)
    count = len(ips) if ips and qtype == 1 else 0
    header = struct.pack("!HHHHHH", qid, 0x8580 | rcode, 1, count, 0, 0)
    return header + question + answers


def main():
    load()
    signal.signal(signal.SIGHUP, load)
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(("10.20.0.2", 53))
    print("DNS interno escutando em 10.20.0.2:53", flush=True)
    while True:
        data, addr = sock.recvfrom(512)
        try:
            resp = answer(data)
        except Exception as exc:  # consulta malformada
            print(f"consulta inválida de {addr}: {exc}", file=sys.stderr, flush=True)
            continue
        if resp:
            sock.sendto(resp, addr)


if __name__ == "__main__":
    main()
PY

cat > /etc/systemd/system/dns-interno.service <<'EOT'
[Unit]
Description=DNS interno (zona "interno")
After=rede-interna.service
Requires=rede-interna.service

[Service]
ExecStart=/usr/bin/python3 /opt/dns-interno/dns.py
ExecReload=/bin/kill -HUP $MAINPID
Restart=always

[Install]
WantedBy=multi-user.target
EOT

# O servidor usa o DNS interno
cat > /etc/resolv.conf <<'EOT'
# DNS interno
nameserver 10.20.0.2
search interno
EOT

# ---------- API de pagamentos (servidor pag-api-01) ----------
cat > /etc/pagamentos.conf <<'EOT'
# API de pagamentos — endereço do servidor atual
BIND=10.20.0.9
PORT=9000
EOT

cat > /opt/pagamentos/api.py <<'PY'
#!/usr/bin/env python3
"""API de pagamentos (simulada): autoriza qualquer pedido."""
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

conf = dict(l.strip().split("=", 1) for l in open("/etc/pagamentos.conf")
            if "=" in l and not l.lstrip().startswith("#"))


class H(BaseHTTPRequestHandler):
    def do_POST(self):
        body = json.dumps({"autorizado": True, "servidor": conf["BIND"]}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *a):
        pass


print(f"API de pagamentos em {conf['BIND']}:{conf['PORT']}", flush=True)
ThreadingHTTPServer((conf["BIND"], int(conf["PORT"])), H).serve_forever()
PY

cat > /etc/systemd/system/pagamentos-api.service <<'EOT'
[Unit]
Description=API de pagamentos (servidor simulado)
After=rede-interna.service
Requires=rede-interna.service

[Service]
ExecStart=/usr/bin/python3 /opt/pagamentos/api.py
Restart=always

[Install]
WantedBy=multi-user.target
EOT

# ---------- Loja Exemplo ----------
cat > /etc/app/app.conf <<'EOT'
# Configuração da Loja Exemplo
PAGAMENTOS_URL=http://api.pagamentos.interno:9000/autorizar
PAGAMENTOS_TIMEOUT=3
EOT

cat > /opt/app/app.py <<'PY'
#!/usr/bin/env python3
"""Loja Exemplo: checkout que depende da API de pagamentos."""
import json
import socket
import urllib.parse
from http.client import HTTPConnection
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

conf = dict(l.strip().split("=", 1) for l in open("/etc/app/app.conf")
            if "=" in l and not l.lstrip().startswith("#"))
url = urllib.parse.urlparse(conf["PAGAMENTOS_URL"])
TIMEOUT = float(conf.get("PAGAMENTOS_TIMEOUT", "3"))

# "Otimização" antiga: resolve o nome uma vez, na inicialização, e guarda o IP.
PAG_HOST, PAG_PORT = url.hostname, url.port or 80
PAG_IP = socket.gethostbyname(PAG_HOST)
print(f"pagamentos: {PAG_HOST} -> {PAG_IP} (resolvido na inicialização)", flush=True)


def autorizar():
    conn = HTTPConnection(PAG_IP, PAG_PORT, timeout=TIMEOUT)
    conn.request("POST", url.path, body=b"{}", headers={"Host": PAG_HOST})
    return json.loads(conn.getresponse().read())


class H(BaseHTTPRequestHandler):
    def send(self, code, body, ctype="application/json"):
        data = body.encode() if isinstance(body, str) else body
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        if self.path == "/health":
            return self.send(200, json.dumps({"status": "ok", "pagamentos": PAG_HOST,
                                              "pagamentos_ip": PAG_IP}))
        if self.path.startswith("/checkout"):
            try:
                r = autorizar()
            except (OSError, ValueError) as exc:
                print(f"ERRO checkout: falha ao chamar {PAG_HOST} ({PAG_IP}:{PAG_PORT}): {exc}", flush=True)
                return self.send(502, json.dumps({"erro": "gateway de pagamento indisponível"}))
            return self.send(200, json.dumps({"pedido": "confirmado", "pagamento": r}))
        self.send(200, "<h1>Loja Exemplo</h1><p>Catálogo funcionando.</p>\n", "text/html; charset=utf-8")

    def log_message(self, *a):
        pass


ThreadingHTTPServer(("127.0.0.1", 8080), H).serve_forever()
PY

cat > /etc/systemd/system/app.service <<'EOT'
[Unit]
Description=Loja Exemplo
After=network.target dns-interno.service pagamentos-api.service

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

# Workaround antigo, de quando o DNS interno estava instável
cat >> /etc/hosts <<'EOT'

# workaround temporário (INC-2291): DNS interno instável, fixando o IP da API
10.20.0.9    api.pagamentos.interno
EOT

systemctl daemon-reload
systemctl enable --now rede-interna dns-interno pagamentos-api
for _ in $(seq 1 20); do
  getent hosts dns.interno >/dev/null && break
  sleep 0.3
done
systemctl enable --now app
nginx -t
systemctl enable nginx
systemctl restart nginx

for _ in $(seq 1 30); do
  curl -fs http://localhost/checkout | grep -q confirmado && exit 0
  sleep 0.5
done
echo "setup: o checkout não respondeu" >&2
exit 1
