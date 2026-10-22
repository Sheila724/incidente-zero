# Para estudar mais

## Kubernetes

- [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/): rolling update, `maxSurge`, `maxUnavailable` e `progressDeadlineSeconds`
- [Imagens](https://kubernetes.io/docs/concepts/containers/images/): nomes, tags, digests e `imagePullPolicy`
- [kubectl rollout](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_rollout/): status, history, undo
- [Depurando pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/)

## Registries

- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md): a API HTTP de uma registry

## Para praticar

- Rode `kubectl rollout history deploy/loja-web` e depois `kubectl rollout undo deploy/loja-web --to-revision=N`. O que acontece com os ReplicaSets?
- Cenários relacionados: **016 — Pods em CrashLoopBackOff** e **018 — Service sem endpoints**.
