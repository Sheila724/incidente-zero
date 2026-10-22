# Para estudar mais

## Kubernetes

- [Gerenciando recursos de containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/): requests, limits e unidades
- [Scheduler do Kubernetes](https://kubernetes.io/docs/concepts/scheduling-eviction/kube-scheduler/)
- [LimitRange](https://kubernetes.io/docs/concepts/policy/limit-range/) e [ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Depurando pods Pending](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/#my-pod-stays-pending)

## Para praticar

- Peça `cpu: 1500m` em cada réplica e escale para 2. Quantas sobem? Leia o evento do scheduler.
- Cenários relacionados: **016 — Pods em CrashLoopBackOff** e **017 — Deploy travado em ImagePullBackOff**.
