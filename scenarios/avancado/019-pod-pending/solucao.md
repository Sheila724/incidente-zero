# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja os pods do checkout

```bash
kubectl get pods
```

```
NAME                             READY   STATUS    RESTARTS   AGE
loja-checkout-6f9b8c7d54-4jq8w   0/1     Pending   0          3m
loja-checkout-6f9b8c7d54-z2m7k   0/1     Pending   0          3m
loja-web-5c8d9f7b6d-7k2lq        1/1     Running   0          25m
loja-web-5c8d9f7b6d-x9p4r        1/1     Running   0          25m
```

Sem reinícios e sem erro: os pods nunca chegaram a rodar. Por isso `kubectl logs` não mostra nada.

## 2. Pergunte ao scheduler

```bash
kubectl describe pod loja-checkout-6f9b8c7d54-4jq8w | tail -4
```

```
Warning  FailedScheduling  default-scheduler  0/1 nodes are available: 1 Insufficient cpu.
         preemption: 0/1 nodes are available: 1 No preemption victims found for incoming pod.
```

O scheduler não achou nenhum nó com CPU livre suficiente para a reserva do pod.

## 3. Compare pedido e oferta

```bash
kubectl get deploy loja-checkout -o jsonpath='{.spec.template.spec.containers[0].resources}'; echo
kubectl describe node loja-k8s-01 | grep -A6 -E 'Allocatable|Allocated resources'
cat /opt/loja/checkout/README.md
```

```
{"limits":{"memory":"64Mi"},"requests":{"cpu":"250","memory":"32Mi"}}

Allocatable:
  cpu:     2        ← núcleos da máquina (o seu pode mostrar outro número)
...
Recursos recomendados por réplica: 250m de CPU e 32Mi de memória.
```

O pod pede **250 núcleos**. O time queria `250m`, um quarto de núcleo. Nenhum nó do mundo atende esse pedido, então nem adianta liberar espaço.

## 4. A armadilha: abrir espaço ou tirar a reserva

- Escalar a `loja-web` para zero não resolve (250 núcleos continuam não cabendo) e derruba a loja principal.
- Remover os `requests` faz o pod subir, mas sem reserva nenhuma: no pico, ele disputa CPU com a loja.

## 5. Corrija a reserva

```bash
kubectl set resources deploy/loja-checkout --requests=cpu=250m
kubectl rollout status deploy/loja-checkout
kubectl get pods
curl -s http://localhost/checkout/
```

```
deployment "loja-checkout" successfully rolled out
<h1>Checkout da Loja Exemplo</h1>
```

Também vale editar com `kubectl edit deploy loja-checkout` e trocar `cpu: "250"` por `cpu: 250m`.
