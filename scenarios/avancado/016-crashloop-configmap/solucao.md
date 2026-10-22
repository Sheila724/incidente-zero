# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja o estado dos pods

```bash
kubectl get pods
```

```
NAME                        READY   STATUS             RESTARTS      AGE
loja-web-7d9c6b7f4d-2xk8p   0/1     CrashLoopBackOff   4 (30s ago)   2m
loja-web-7d9c6b7f4d-q7w2m   0/1     CrashLoopBackOff   4 (25s ago)   2m
```

Nenhuma réplica pronta (`0/1`), e o número de reinícios só aumenta.

## 2. A armadilha: apagar o pod

```bash
kubectl delete pod loja-web-7d9c6b7f4d-2xk8p
kubectl get pods -w
```

O Deployment cria outro pod na hora, que vai para `Error` e depois para `CrashLoopBackOff`. O pod novo é igual ao antigo: mesma imagem, mesma configuração.

## 3. Leia o motivo

```bash
kubectl logs deploy/loja-web --previous
```

```
Loja Exemplo 2.5 iniciando...
ERRO FATAL: variável de ambiente DATABASE_URL não definida
Não é possível conectar ao banco. Encerrando.
```

`--previous` mostra a saída da execução que **morreu**. Sem ele, você pode pegar o container novo, que ainda não escreveu nada.

## 4. Siga a configuração

```bash
kubectl get deploy loja-web -o yaml | grep -A3 envFrom
kubectl get configmap loja-config -o yaml
```

```yaml
envFrom:
- configMapRef:
    name: loja-config
---
data:
  CACHE_TTL: "300"
  DB_URL: mysql://loja@db.loja.svc:3306/loja
metadata:
  annotations:
    loja.local/alterado-em: "ontem, 18:40"
    loja.local/motivo: Padronização dos nomes das variáveis
```

A aplicação espera `DATABASE_URL`, mas a chave foi renomeada para `DB_URL` ontem. Os pods antigos já tinham a variável carregada e continuaram funcionando. Quando a manutenção recriou os pods, eles subiram com a configuração nova e morreram.

## 5. Corrija

Crie de novo a chave que a aplicação espera:

```bash
kubectl patch configmap loja-config --type merge \
  -p '{"data":{"DATABASE_URL":"mysql://loja@db.loja.svc:3306/loja"}}'
```

(ou `kubectl edit configmap loja-config` e adicione a linha). Manter `DB_URL` também não atrapalha e não quebra quem já usa o nome novo.

## 6. Recrie os pods

Os pods em `CrashLoopBackOff` podem levar minutos para tentar de novo, e só leem a variável na subida:

```bash
kubectl rollout restart deploy/loja-web
kubectl rollout status deploy/loja-web
kubectl get pods
curl -s http://localhost/health
```

```
deployment "loja-web" successfully rolled out
ok versao=2.5 db=mysql://loja@db.loja.svc:3306/loja
```

## Outra correção válida

Em vez de mudar a ConfigMap, dá para mapear o nome novo no Deployment:

```yaml
env:
- name: DATABASE_URL
  valueFrom:
    configMapKeyRef: {name: loja-config, key: DB_URL}
```

O importante é que a aplicação e a configuração voltem a concordar, e que isso fique registrado.
