#!/usr/bin/env python3
"""Registry interna da Loja Exemplo (só para o lab).

Implementa o mínimo da API de registry (OCI distribution) para o containerd
e para quem investiga com curl:
  GET  /v2/                      -> 200
  GET  /v2/_catalog              -> repositórios
  GET  /v2/<repo>/tags/list      -> tags de /srv/registry/<repo>.tags
  HEAD/GET /v2/<repo>/manifests/<tag> -> 404 MANIFEST_UNKNOWN
As imagens publicadas já estão carregadas no nó, então o containerd só
consulta a registry quando a tag pedida não existe.
"""
import json
import os
import re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

DIR = "/srv/registry"


def repos():
    return sorted(f[:-5] for f in os.listdir(DIR) if f.endswith(".tags"))


def tags(repo):
    try:
        with open(os.path.join(DIR, repo + ".tags")) as f:
            return [t.strip() for t in f if t.strip()]
    except OSError:
        return None


class Registry(BaseHTTPRequestHandler):
    server_version = "registry-loja/1.0"

    def responder(self, code, corpo):
        dados = json.dumps(corpo).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Docker-Distribution-API-Version", "registry/2.0")
        self.send_header("Content-Length", str(len(dados)))
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(dados)

    def erro(self, code, codigo, msg, detalhe=None):
        self.responder(code, {"errors": [{"code": codigo, "message": msg, "detail": detalhe}]})

    def do_GET(self):
        path = self.path.split("?")[0]
        if path in ("/v2", "/v2/"):
            return self.responder(200, {})
        if path == "/v2/_catalog":
            return self.responder(200, {"repositories": repos()})
        m = re.fullmatch(r"/v2/(.+)/tags/list", path)
        if m:
            t = tags(m.group(1))
            if t is None:
                return self.erro(404, "NAME_UNKNOWN", "repository name not known to registry", {"name": m.group(1)})
            return self.responder(200, {"name": m.group(1), "tags": t})
        m = re.fullmatch(r"/v2/(.+)/manifests/(.+)", path)
        if m:
            if tags(m.group(1)) is None:
                return self.erro(404, "NAME_UNKNOWN", "repository name not known to registry", {"name": m.group(1)})
            return self.erro(404, "MANIFEST_UNKNOWN", "manifest unknown", {"Tag": m.group(2)})
        return self.erro(404, "NOT_FOUND", "not found")

    do_HEAD = do_GET

    def log_message(self, fmt, *args):
        print("%s %s" % (self.address_string(), fmt % args), flush=True)


if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", 5000), Registry).serve_forever()
