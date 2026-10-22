# Conceitos para este cenário

## O caminho de um pod até rodar

```
kubectl apply → API aceita → Pending → scheduler escolhe um nó → kubelet cria o container → Running
```

`Pending` é o pod que já existe na API, mas ainda não tem nó (ou ainda está baixando a imagem). Como nenhum container rodou, **não há logs**: a explicação está nos eventos.

```bash
kubectl describe pod <pod>     # Events, no final
kubectl get events --sort-by=.lastTimestamp
```

## Requests e limits

```yaml
resources:
  requests: {cpu: 250m, memory: 64Mi}   # o que o scheduler reserva no nó
  limits:   {memory: 128Mi}             # o teto que o container pode usar
```

| Campo | Para que serve |
|---|---|
| **requests** | O scheduler só coloca o pod num nó que tenha esse tanto **livre para reservar**. É uma reserva, não o uso real |
| **limits** | O máximo que o container pode usar. Passou da memória: o container é encerrado (OOMKilled). Passou da CPU: é freado |

## Unidades

| Recurso | Exemplos |
|---|---|
| CPU | `1` = um núcleo · `500m` = meio núcleo · `100m` = um décimo |
| Memória | `64Mi` = 64 mebibytes · `1Gi` = 1 gibibyte · `64M` = 64 megabytes |

Um `m` a menos muda o pedido em mil vezes.

## Quanto o nó oferece

```bash
kubectl describe node loja-k8s-01
```

- **Capacity:** o total da máquina.
- **Allocatable:** o que sobra para os pods depois das reservas do sistema.
- **Allocated resources:** a soma dos requests de todos os pods que já estão no nó.

O scheduler compara o pedido do pod com `Allocatable − Allocated`.

## Por que não remover os requests

Sem requests, o scheduler coloca o pod em qualquer nó, mesmo que ele já esteja no limite. O pod sobe, mas disputa CPU e memória com os outros e pode derrubar vizinhos em horário de pico. A reserva certa é a que protege.
