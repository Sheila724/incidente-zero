#!/usr/bin/env bash
# Cenário 019 — injeta a falha:
# primeiro deploy do loja-checkout com requests.cpu "250" (250 núcleos) em vez
# de "250m". Os pods ficam em Pending e /checkout/ responde 502.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl -n loja apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-checkout-app
data:
  start.sh: |
    #!/bin/sh
    echo "loja-checkout 1.0 iniciando..."
    mkdir -p /www/checkout
    printf '<h1>Checkout da Loja Exemplo</h1>\n<p>Versão 1.0</p>\n' > /www/checkout/index.html
    echo ok > /www/checkout/health
    echo "Pronto na porta 8080"
    exec httpd -f -p 8080 -h /www
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loja-checkout
  labels: {app: loja-checkout}
  annotations:
    kubernetes.io/change-cause: "pipeline: primeiro deploy do checkout 1.0"
spec:
  replicas: 2
  selector:
    matchLabels: {app: loja-checkout}
  template:
    metadata:
      labels: {app: loja-checkout}
    spec:
      containers:
        - name: checkout
          image: registry.loja.local/loja-checkout:1.0
          command: ["sh", "/app/start.sh"]
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet: {path: /checkout/health, port: 8080}
            periodSeconds: 3
          resources:
            requests: {cpu: "250", memory: 32Mi}
            limits: {memory: 64Mi}
          volumeMounts:
            - {name: app, mountPath: /app}
      volumes:
        - name: app
          configMap: {name: loja-checkout-app}
---
apiVersion: v1
kind: Service
metadata:
  name: loja-checkout
spec:
  type: NodePort
  selector: {app: loja-checkout}
  ports:
    - {port: 80, targetPort: 8080, nodePort: 30081}
YAML

# Autoverificação: pods em Pending por falta de CPU e /checkout/ com 502
for _ in $(seq 1 60); do
  pendentes=$(kubectl -n loja get pods -l app=loja-checkout --field-selector=status.phase=Pending -o name 2>/dev/null | wc -l)
  evento=$(kubectl -n loja get events --field-selector reason=FailedScheduling -o jsonpath='{.items[*].message}' 2>/dev/null)
  code=$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/checkout/)
  if [ "$pendentes" -ge 2 ] && grep -q 'Insufficient cpu' <<< "$evento" && [ "$code" = "502" ]; then
    exit 0
  fi
  sleep 2
done
echo "inject: os pods do checkout não ficaram em Pending como esperado" >&2
exit 1
