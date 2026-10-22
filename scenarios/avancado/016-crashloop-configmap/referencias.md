# Para estudar mais

## Kubernetes

- [Depurando pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/): o roteiro oficial de investigação
- [Ciclo de vida do pod](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/): estados, reinícios e o backoff
- [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/): por que variáveis de ambiente não atualizam sozinhas
- [kubectl rollout](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_rollout/): restart, status e undo
- [Cheat sheet do kubectl](https://kubernetes.io/docs/reference/kubectl/quick-reference/)

## Para praticar

- Monte a ConfigMap como arquivo (volume) em vez de variáveis. O que acontece com o arquivo dentro do pod quando a ConfigMap muda?
- Cenários relacionados: **011 — 502 Bad Gateway após deploy** e **017 — Deploy travado em ImagePullBackOff**.
