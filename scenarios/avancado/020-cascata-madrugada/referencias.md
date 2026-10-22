# Para estudar mais

## Kubernetes

- [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) e [Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Labels e seletores](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/): separando ambientes
- [Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/): uma base, vários ambientes
- [Depurando Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/)

## Incidentes

- [Google SRE Book: Managing Incidents](https://sre.google/sre-book/managing-incidents/): mitigar antes de investigar
- [Google SRE Book: Addressing Cascading Failures](https://sre.google/sre-book/addressing-cascading-failures/)

## Para praticar

- Rode `kubectl get deploy -l ambiente=homologacao` antes e depois da correção. Quem entra e quem sai do alcance da automação?
- Cenários relacionados: **018 — Service sem endpoints** e **016 — Pods em CrashLoopBackOff**.
