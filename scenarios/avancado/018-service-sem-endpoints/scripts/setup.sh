#!/usr/bin/env bash
# Cenário 018 — prepara o ambiente SAUDÁVEL:
#   k3s de nó único; Deployment loja-web (2 réplicas, versão 2.6, estratégia
#   Recreate) com readiness probe em /health; Service NodePort 30080;
#   nginx do servidor (80) -> Service.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

k3s-lab-up 2.4 2.5 2.6

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
      echo "ERRO FATAL: variável de ambiente DATABASE_URL não definida" >&2
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
  strategy:
    type: Recreate   # a subida migra o banco: duas versões não podem rodar juntas
  selector:
    matchLabels: {app: loja-web}
  template:
    metadata:
      labels: {app: loja-web}
    spec:
      containers:
        - name: web
          image: registry.loja.local/loja-web:2.6
          command: ["sh", "/app/start.sh"]
          env:
            - {name: APP_VERSION, value: "2.6"}
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
for _ in $(seq 1 30); do
  [ "$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)" = "200" ] && exit 0
  sleep 1
done
echo "setup: a loja não respondeu 200" >&2
exit 1
