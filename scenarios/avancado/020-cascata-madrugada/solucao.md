# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Siga o caminho do pedido

```bash
curl -I http://localhost/
kubectl get pods
kubectl describe svc loja-web | grep Endpoints
```

```
HTTP/1.1 502 Bad Gateway

NAME                                  READY   STATUS      RESTARTS   AGE
economia-homologacao-29878252-r7rnr   0/1     Completed   0          40s
loja-web-7bcb76664b-nfvls             0/1     Running     0          3h
loja-web-7bcb76664b-wg8gv             0/1     Running     0          3h

Endpoints:
```

O nginx está de pé, o Service não tem endpoints e os pods da loja estão `Running`, mas não prontos. Os pods `economia-homologacao-...` em `Completed` são execuções de um Job: guarde essa pista.

## 2. Por que a loja não está pronta?

```bash
kubectl describe pod loja-web-7bcb76664b-nfvls | grep Unhealthy
```

```
Warning  Unhealthy  kubelet  Readiness probe failed: banco loja-db:3306 inacessível
```

A loja depende do banco, e o banco não responde.

## 3. Restaure primeiro

```bash
kubectl get deploy
```

```
NAME             READY   UP-TO-DATE   AVAILABLE
loja-db          0/0     0            0
loja-web         0/2     2            0
relatorios-web   0/0     0            0
```

O banco está com **zero réplicas desejadas**: não caiu, foi desligado. Antes de investigar quem fez isso, devolva a loja aos clientes:

```bash
kubectl scale deploy/loja-db --replicas=1
kubectl rollout status deploy/loja-web
curl -I http://localhost/
```

A loja volta. Se você clicar em **Verificar** agora, quase tudo passa, mas não tudo.

## 4. Espere e observe

```bash
kubectl get deploy loja-db -w
```

Em até 2 minutos, o banco volta para `0/0` e a loja cai de novo. Alguém está desligando o banco:

```bash
kubectl get events --sort-by=.lastTimestamp | grep -i scaled
kubectl get cronjobs
kubectl get jobs
```

```
Scaled down replica set loja-db-7cdfb4854 from 1 to 0

NAME                   SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE
economia-homologacao   */2 * * * *   False     0        40s
```

## 5. Ache a origem

```bash
kubectl logs job/economia-homologacao-29878254
kubectl get deploy --show-labels
cat /opt/loja/infra/CHANGELOG.md
```

```
economia-homologacao: procurando deployments com ambiente=homologacao e réplicas > 0
desligando loja-db (custo de homologação fora do horário de testes)

NAME             ... LABELS
loja-db          ... ambiente=homologacao,app=loja-db,time=dados
loja-web         ... ambiente=producao,app=loja-web
relatorios-web   ... ambiente=homologacao,app=relatorios-web,time=dados
```

A automação está certa: ela desliga o que está marcado como homologação. O erro é o banco de produção estar marcado assim. O manifesto foi copiado de homologação, e só a imagem e as réplicas foram ajustadas.

## 6. Corrija a causa

```bash
kubectl label deploy/loja-db ambiente=producao --overwrite
kubectl scale deploy/loja-db --replicas=1
kubectl rollout status deploy/loja-web
```

Espere mais 2 minutos e confira: o próximo job diz `nada para desligar` e a loja continua no ar.

## As armadilhas

- **Só subir o banco:** a loja volta e cai de novo no próximo ciclo da CronJob.
- **Suspender ou apagar a CronJob:** a loja fica de pé, mas homologação volta a gastar e o rótulo errado continua lá, esperando a próxima automação que confie nele.
