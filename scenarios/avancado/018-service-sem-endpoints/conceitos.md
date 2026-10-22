# Conceitos para este cenário

## Running não é o mesmo que pronto

```
NAME                       READY   STATUS    RESTARTS
loja-web-6b8f7d9c5-4kq2x   0/1     Running   0
```

| Coluna | O que diz |
|---|---|
| **STATUS** `Running` | O container foi criado e o processo está rodando |
| **READY** `0/1` | Nenhum container do pod passou na readiness probe |

## As probes

| Probe | Se falhar... |
|---|---|
| **readinessProbe** | O pod sai do Service (para de receber tráfego), mas **não** é reiniciado |
| **livenessProbe** | O container é reiniciado |
| **startupProbe** | Segura as outras probes enquanto a aplicação inicia |

Uma readiness probe HTTP considera sucesso qualquer código de **200 a 399**. Um 404 é falha.

```yaml
readinessProbe:
  httpGet: {path: /health, port: 8080}
  periodSeconds: 3
```

## Como o Service escolhe os pods

O Service tem um **selector** (ex.: `app: loja-web`). O Kubernetes mantém uma lista de **endpoints**: os IPs dos pods que batem com o selector **e estão prontos**.

```bash
kubectl describe svc loja-web            # linha "Endpoints:"
kubectl get endpointslices -l kubernetes.io/service-name=loja-web
```

Se a lista está vazia, o Service não tem para onde mandar o tráfego, e quem está na frente (o nginx) responde 502.

## Testando de dentro do pod

```bash
kubectl exec <pod> -- wget -qO- http://localhost:8080/health
kubectl exec <pod> -- wget -S -qO- http://localhost:8080/outro-caminho
```

## Por que não remover a probe?

Sem readiness probe, o pod entra no Service assim que o container inicia, antes de a aplicação estar pronta. Em todo deploy, uma parte dos clientes recebe erro. A probe errada derrubou a loja; a probe certa a protege.
