# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja o rollout

```bash
kubectl rollout status deploy/loja-web
kubectl get pods
```

```
error: deployment "loja-web" exceeded its progress deadline

NAME                        READY   STATUS             RESTARTS   AGE
loja-web-5c8d9f7b6d-7k2lq   1/1     Running            0          20m
loja-web-5c8d9f7b6d-x9p4r   1/1     Running            0          20m
loja-web-84f6c9d5b8-m2tqz   0/1     ImagePullBackOff   0          3m
```

## 2. Por que a loja não caiu

```bash
kubectl get rs
```

```
NAME                  DESIRED   CURRENT   READY
loja-web-5c8d9f7b6d   2         2         2     ← versão 2.5, servindo
loja-web-84f6c9d5b8   1         1         0     ← versão 2.6, travada
```

O rolling update cria um pod novo e só remove um antigo quando o novo fica pronto. Como o novo nunca ficou, os dois antigos continuam atendendo.

## 3. Leia o motivo

```bash
kubectl describe pod loja-web-84f6c9d5b8-m2tqz | tail -8
```

```
Failed   Failed to pull image "registry.loja.local/loja-web:v2.6": ...
         registry.loja.local/loja-web:v2.6: not found
Warning  Failed   Error: ImagePullBackOff
```

A registry respondeu, então ela está no ar. O que não existe é a imagem `loja-web:v2.6`.

## 4. Consulte a registry e o pipeline

```bash
curl -s http://registry.loja.local:5000/v2/loja-web/tags/list
cat /var/log/pipeline/deploy-2.6.log
```

```
{"name": "loja-web", "tags": ["2.4", "2.5", "2.6"]}

[build] docker push registry.loja.local/loja-web:2.6
[deploy] kubectl set image deploy/loja-web web=registry.loja.local/loja-web:v2.6
```

O build publicou `2.6`; o deploy pediu `v2.6`, que é o nome da tag do Git.

## 5. Conclua o deploy

```bash
kubectl set image deploy/loja-web web=registry.loja.local/loja-web:2.6
kubectl rollout status deploy/loja-web
kubectl get pods
curl -s http://localhost/health
```

```
deployment "loja-web" successfully rolled out
ok versao=2.6 db=mysql://loja@db.loja.svc:3306/loja
```

O pod travado é substituído e os antigos saem um a um, conforme os novos ficam prontos.

## A armadilha: desfazer

`kubectl rollout undo deploy/loja-web` também "resolve" o alerta do pipeline, mas a promoção não vai ao ar e o erro do pipeline continua lá para o próximo deploy.
