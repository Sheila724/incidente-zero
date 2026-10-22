# Conceitos para este cenário

## As peças da loja no Kubernetes

| Objeto | Papel |
|---|---|
| **Pod** | Um ou mais containers rodando juntos. É descartável: morre e é recriado |
| **Deployment** | Garante que existam N pods iguais (réplicas) e cuida das atualizações |
| **ConfigMap** | Configuração (chave = valor) separada da imagem |
| **Service** | Endereço estável que distribui o tráfego entre os pods prontos |

```
cliente → nginx :80 → Service loja-web :30080 → pods loja-web (2 réplicas)
```

## Comandos que você vai usar

```bash
kubectl get pods                    # estado e reinícios de cada pod
kubectl describe pod <pod>          # detalhes e eventos (o "histórico" do pod)
kubectl logs <pod>                  # saída do container atual
kubectl logs <pod> --previous       # saída da execução que morreu
kubectl get deploy loja-web -o yaml # a especificação completa
kubectl get events --sort-by=.lastTimestamp
```

## Estados de um pod que valem decorar

| STATUS | O que significa |
|---|---|
| `Running` com `READY 1/1` | Container rodando e pronto para receber tráfego |
| `CrashLoopBackOff` | O container inicia e morre; o Kubernetes espera cada vez mais para tentar de novo (10 s, 20 s, 40 s... até 5 min) |
| `Error` | O container terminou com erro (código de saída diferente de 0) |
| `CreateContainerConfigError` | Falta algo para montar o container (ex.: uma ConfigMap ou Secret que não existe) |

## Por que apagar o pod não resolve

O Deployment cria um pod novo, **idêntico**, com a mesma imagem e a mesma configuração. Se o problema está na configuração, o pod novo morre do mesmo jeito.

## Configuração só é lida na subida

Variáveis de ambiente vindas de uma ConfigMap (`env` ou `envFrom`) são copiadas para o container **quando ele inicia**. Mudar a ConfigMap não afeta containers que já estão rodando, e é por isso que uma mudança errada pode ficar "dormindo" até o próximo reinício.

Para recriar os pods de um Deployment de forma controlada:

```bash
kubectl rollout restart deploy/loja-web
kubectl rollout status deploy/loja-web
```
