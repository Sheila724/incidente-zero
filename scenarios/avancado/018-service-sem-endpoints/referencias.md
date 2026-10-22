# Para estudar mais

## Kubernetes

- [Liveness, readiness e startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Service](https://kubernetes.io/docs/concepts/services-networking/service/): selector e endpoints
- [EndpointSlices](https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/)
- [Depurando Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/): o roteiro oficial, passo a passo
- [Estratégias de Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy): Recreate e RollingUpdate

## Para praticar

- Troque a estratégia para RollingUpdate e repita a mudança da probe. O que acontece com a loja?
- Cenários relacionados: **011 — 502 Bad Gateway após deploy** e **016 — Pods em CrashLoopBackOff**.
