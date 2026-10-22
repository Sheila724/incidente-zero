#!/usr/bin/env bash
# Cenário 019 — prepara o ambiente SAUDÁVEL:
#   k3s de nó único; loja-web (2 réplicas) no NodePort 30080;
#   registry com a imagem loja-checkout:1.0, ainda sem deploy;
#   nginx do servidor: /checkout/ -> NodePort 30081, o resto -> 30080.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

k3s-lab-up 2.4 2.5

# Imagem do checkout publicada na registry interna
k3s ctr -n k8s.io images tag --force docker.io/rancher/mirrored-library-busybox:1.36.1 \
  registry.loja.local/loja-checkout:1.0 >/dev/null
echo 1.0 > /srv/registry/loja-checkout.tags

kubectl -n loja apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-config
data:
  DATABASE_URL: mysql://loja@db.loja.svc:3306/loja
  CACHE_TTL: "300"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-web-app
data:
  start.sh: |
    #!/bin/sh
    echo "Loja Exemplo $APP_VERSION iniciando..."
    if [ -z "$DATABASE_URL" ]; then
      # O motivo também vai para a mensagem de término (kubectl describe pod → Last State)
      echo "ERRO FATAL: variável de ambiente DATABASE_URL não definida" | tee /dev/termination-log >&2
      echo "Não é possível conectar ao banco. Encerrando." >&2
      exit 1
    fi
    echo "Banco: $DATABASE_URL"
    mkdir -p /www
    printf '<h1>Loja Exemplo</h1>\n<p>Versão %s</p>\n' "$APP_VERSION" > /www/index.html
    printf 'ok versao=%s db=%s\n' "$APP_VERSION" "$DATABASE_URL" > /www/health
    echo "Pronta na porta 8080"
    exec httpd -f -p 8080 -h /www
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loja-web
  labels: {app: loja-web}
spec:
  replicas: 2
  selector:
    matchLabels: {app: loja-web}
  template:
    metadata:
      labels: {app: loja-web}
    spec:
      containers:
        - name: web
          image: registry.loja.local/loja-web:2.5
          command: ["sh", "/app/start.sh"]
          env:
            - {name: APP_VERSION, value: "2.5"}
          envFrom:
            - configMapRef: {name: loja-config}
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet: {path: /health, port: 8080}
            periodSeconds: 3
          resources:
            requests: {cpu: 50m, memory: 16Mi}
            limits: {memory: 64Mi}
          volumeMounts:
            - {name: app, mountPath: /app}
      volumes:
        - name: app
          configMap: {name: loja-web-app}
---
apiVersion: v1
kind: Service
metadata:
  name: loja-web
spec:
  type: NodePort
  selector: {app: loja-web}
  ports:
    - {port: 80, targetPort: 8080, nodePort: 30080}
YAML

kubectl -n loja rollout status deploy/loja-web --timeout=180s >/dev/null

cat > /etc/nginx/sites-available/loja <<'NGX'
# Loja Exemplo: o tráfego entra pelo nginx e vai para os Services do Kubernetes
server {
    listen 80 default_server;
    server_name _;
    location /checkout/ {
        proxy_pass http://127.0.0.1:30081;
        proxy_set_header Host $host;
        proxy_connect_timeout 3s;
    }
    location / {
        proxy_pass http://127.0.0.1:30080;
        proxy_set_header Host $host;
        proxy_connect_timeout 3s;
    }
}
NGX
systemctl reload nginx

install -d -m 0755 /opt/loja/checkout
cat > /opt/loja/checkout/README.md <<'MD'
# loja-checkout

Serviço de checkout da Loja Exemplo (primeira versão: 1.0).

- Porta do container: 8080 · Service NodePort: 30081
- Health check: GET /checkout/health
- Réplicas em produção: 2

## Recursos recomendados por réplica

Medidos no teste de carga de 14/10: **250m de CPU** e **32Mi de memória**
(pico de 61Mi; limite de memória em 64Mi).
MD
for _ in $(seq 1 30); do
  [ "$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)" = "200" ] && exit 0
  sleep 1
done
echo "setup: a loja não respondeu 200" >&2
exit 1
