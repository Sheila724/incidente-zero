#!/usr/bin/env bash
# Cenário 020 — prepara o ambiente SAUDÁVEL:
#   loja-web (2 réplicas) depende do loja-db (readiness probe testa o banco);
#   relatorios-web é um ambiente de homologação, desligado;
#   CronJob economia-homologacao desliga, a cada 2 minutos, tudo com
#   ambiente=homologacao. nginx (80) -> Service loja-web (NodePort 30080).
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

k3s-lab-up 2.4 2.5 2.6

base=docker.io/rancher/mirrored-library-busybox:1.36.1
for img in loja-db:10.11 relatorios-web:0.9 ops-tools:1.0; do
  k3s ctr -n k8s.io images tag --force "$base" "registry.loja.local/$img" >/dev/null
  echo "${img#*:}" > "/srv/registry/${img%%:*}.tags"
done

kubectl -n loja apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-web-app
data:
  start.sh: |
    #!/bin/sh
    echo "Loja Exemplo $APP_VERSION iniciando..."
    mkdir -p /www
    printf '<h1>Loja Exemplo</h1>\n<p>Versão %s</p>\n' "$APP_VERSION" > /www/index.html
    printf 'ok versao=%s\n' "$APP_VERSION" > /www/health
    exec httpd -f -p 8080 -h /www
  pronta.sh: |
    #!/bin/sh
    # A loja só recebe clientes se o banco responde
    wget -q -T 2 -O /dev/null http://loja-db:3306/ 2>/dev/null || {
      echo "banco loja-db:3306 inacessível"
      exit 1
    }
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: loja-db-app
data:
  start.sh: |
    #!/bin/sh
    echo "loja-db (MariaDB 10.11 de laboratório) aceitando conexões na porta 3306"
    mkdir -p /srv/db && echo "loja-db ok" > /srv/db/index.html
    exec httpd -f -p 3306 -h /srv/db
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loja-web
  labels: {app: loja-web, ambiente: producao}
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
          image: registry.loja.local/loja-web:2.6
          command: ["sh", "/app/start.sh"]
          env:
            - {name: APP_VERSION, value: "2.6"}
          ports:
            - containerPort: 8080
          readinessProbe:
            exec: {command: ["sh", "/app/pronta.sh"]}
            periodSeconds: 3
            timeoutSeconds: 3
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loja-db
  labels: {app: loja-db, ambiente: producao, time: dados}
spec:
  replicas: 1
  selector:
    matchLabels: {app: loja-db}
  template:
    metadata:
      labels: {app: loja-db}
    spec:
      containers:
        - name: db
          image: registry.loja.local/loja-db:10.11
          command: ["sh", "/app/start.sh"]
          ports:
            - containerPort: 3306
          resources:
            requests: {cpu: 50m, memory: 16Mi}
            limits: {memory: 64Mi}
          volumeMounts:
            - {name: app, mountPath: /app}
      volumes:
        - name: app
          configMap: {name: loja-db-app}
---
apiVersion: v1
kind: Service
metadata:
  name: loja-db
spec:
  selector: {app: loja-db}
  ports:
    - {port: 3306, targetPort: 3306}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: relatorios-web
  labels: {app: relatorios-web, ambiente: homologacao, time: dados}
spec:
  replicas: 0
  selector:
    matchLabels: {app: relatorios-web}
  template:
    metadata:
      labels: {app: relatorios-web}
    spec:
      containers:
        - name: web
          image: registry.loja.local/relatorios-web:0.9
          command: ["httpd", "-f", "-p", "8080", "-h", "/tmp"]
          resources:
            requests: {cpu: 50m, memory: 16Mi}
---
# Automação de custo: ambientes de homologação só ficam ligados quando
# alguém sobe para testar; a cada 2 minutos, tudo com ambiente=homologacao
# volta para zero réplicas.
apiVersion: v1
kind: ServiceAccount
metadata:
  name: economia
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: economia
rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "deployments/scale"]
    verbs: ["get", "list", "patch", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: economia
subjects:
  - {kind: ServiceAccount, name: economia}
roleRef:
  {apiGroup: rbac.authorization.k8s.io, kind: Role, name: economia}
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: economia-homologacao
  labels: {time: plataforma}
  annotations:
    plataforma.loja.local/descricao: "Desliga os ambientes de homologação (ambiente=homologacao) que ficaram ligados"
spec:
  schedule: "*/2 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      backoffLimit: 0
      template:
        spec:
          serviceAccountName: economia
          restartPolicy: Never
          containers:
            - name: economia
              image: registry.loja.local/ops-tools:1.0
              command:
                - sh
                - -c
                - |
                  echo "economia-homologacao: procurando deployments com ambiente=homologacao e réplicas > 0"
                  alvos=$(kubectl get deploy -l ambiente=homologacao -o jsonpath='{range .items[?(@.spec.replicas>0)]}{.metadata.name}{" "}{end}')
                  if [ -z "$alvos" ]; then echo "nada para desligar"; exit 0; fi
                  for d in $alvos; do
                    echo "desligando $d (custo de homologação fora do horário de testes)"
                    kubectl scale deploy "$d" --replicas=0
                  done
              # Em produção, seria uma imagem com kubectl. No lab, o kubectl vem do nó.
              volumeMounts:
                - {name: kubectl, mountPath: /usr/local/bin/kubectl, readOnly: true}
                - {name: k3s-data, mountPath: /var/lib/rancher/k3s/data}
          volumes:
            - name: kubectl
              hostPath: {path: /usr/local/bin/k3s, type: File}
            - name: k3s-data
              hostPath: {path: /var/lib/rancher/k3s/data, type: Directory}
YAML

kubectl -n loja rollout status deploy/loja-db --timeout=180s >/dev/null
kubectl -n loja rollout status deploy/loja-web --timeout=180s >/dev/null

install -d -m 0755 /opt/loja/infra
cat > /opt/loja/infra/CHANGELOG.md <<'MD'
# Mudanças de infraestrutura — cluster loja-k8s-01

## Ontem
- 18:30 — time de dados: `loja-db` com manifesto novo (MariaDB 10.11), o mesmo
  testado em homologação. "Só trocamos a imagem e a quantidade de réplicas."

## Semana passada
- CronJob `economia-homologacao` (time de plataforma): a cada 2 minutos, desliga
  os ambientes de homologação que ficaram ligados. Economia estimada: 30%.
- `relatorios-web` (homologação) criado para os testes do time de dados.
MD

for _ in $(seq 1 30); do
  [ "$(curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost/)" = "200" ] && exit 0
  sleep 1
done
echo "setup: a loja não respondeu 200" >&2
exit 1
