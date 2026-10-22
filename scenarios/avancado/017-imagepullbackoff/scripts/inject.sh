#!/usr/bin/env bash
# Cenário 017 — injeta a falha:
# o build publicou loja-web:2.6, mas o deploy aplicou loja-web:v2.6 (a tag do Git).
# O rolling update fica travado com um pod em ImagePullBackOff.
set -euo pipefail
umask 022
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

cat > /var/log/pipeline/deploy-2.6.log <<'LOG'
[pipeline] loja-web — tag do Git: v2.6
[build]  docker build -t registry.loja.local/loja-web:2.6 .
[build]  docker push registry.loja.local/loja-web:2.6
[build]  2.6: digest: sha256:9d1f07a3b5c4 size: 527
[deploy] IMAGEM=registry.loja.local/loja-web:${GIT_TAG}
[deploy] kubectl set image deploy/loja-web web=registry.loja.local/loja-web:v2.6
[deploy] kubectl rollout status deploy/loja-web --timeout=60s
[deploy] Waiting for deployment "loja-web" rollout to finish: 1 out of 2 new replicas have been updated...
[deploy] error: deployment "loja-web" exceeded its progress deadline
[pipeline] FALHOU (exit 1)
LOG

kubectl -n loja patch deploy loja-web --type json -p '[
  {"op":"replace","path":"/spec/template/spec/containers/0/image","value":"registry.loja.local/loja-web:v2.6"},
  {"op":"replace","path":"/spec/template/spec/containers/0/env/0/value","value":"2.6"},
  {"op":"replace","path":"/metadata/annotations/kubernetes.io~1change-cause","value":"pipeline: deploy 2.6"}
]' >/dev/null

# Autoverificação: pod novo sem imagem e o rollout estourando o prazo
for _ in $(seq 1 75); do
  motivo=$(kubectl -n loja get pods -l app=loja-web \
    -o jsonpath='{range .items[*]}{.status.containerStatuses[0].state.waiting.reason}{"\n"}{end}' 2>/dev/null)
  prazo=$(kubectl -n loja get deploy loja-web \
    -o jsonpath='{.status.conditions[?(@.type=="Progressing")].reason}' 2>/dev/null)
  if grep -qE 'ImagePullBackOff|ErrImagePull' <<< "$motivo" && [ "$prazo" = "ProgressDeadlineExceeded" ]; then
    exit 0
  fi
  sleep 2
done
echo "inject: o rollout não travou como esperado" >&2
exit 1
