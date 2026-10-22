# Conceitos para este cenário

## Falha em cascata

Um serviço cai porque **depende** de outro que caiu. O alerta aparece na ponta (a loja), mas a origem está mais para trás. Siga o caminho:

```
nginx → Service loja-web → pods loja-web → (readiness) → Service loja-db → pod loja-db
```

Em cada etapa, pergunte: este está bem? Se sim, olhe o próximo.

## Primeiro restaurar, depois investigar

Num incidente, a ordem é:

1. **Mitigar:** devolver o serviço aos clientes, mesmo que a causa ainda não esteja clara.
2. **Investigar:** entender por que aconteceu.
3. **Corrigir a causa:** para não acontecer de novo.

Se o problema volta depois da mitigação, isso é uma pista valiosa: algo continua agindo.

## Labels e seletores

Labels são etiquetas `chave=valor` nos objetos. Muitas coisas escolhem "em quem agir" por label:

```bash
kubectl get deploy --show-labels
kubectl get deploy -l ambiente=homologacao      # quem tem este rótulo
kubectl label deploy/<nome> chave=valor --overwrite
```

Um rótulo errado faz um objeto entrar (ou sair) do alcance de um Service, de uma política ou de uma automação.

## CronJobs e Jobs

| Objeto | O que é |
|---|---|
| **CronJob** | Agenda (no formato do cron) que cria Jobs |
| **Job** | Uma execução: roda um pod até terminar |

```bash
kubectl get cronjobs
kubectl get jobs                    # as últimas execuções
kubectl logs job/<nome>             # o que a execução fez
kubectl describe cronjob <nome>     # agenda, último horário, se está suspensa
```

`kubectl patch cronjob <nome> -p '{"spec":{"suspend":true}}'` suspende uma CronJob. É uma mitigação válida no meio do incidente, mas desliga a automação inteira.

## Eventos dizem o quê, não quem

```bash
kubectl get events --sort-by=.lastTimestamp
```

Um evento como `Scaled down replica set loja-db-xxx from 1 to 0` mostra **o que** aconteceu. Para saber **quem** pediu, procure automações (CronJobs, operadores, pipelines) e os logs delas.
