# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Confirme o sintoma

```bash
curl -I http://localhost/
kubectl get pods
```

```
HTTP/1.1 502 Bad Gateway

NAME                        READY   STATUS    RESTARTS   AGE
loja-web-6b8f7d9c5-4kq2x    0/1     Running   0          8m
loja-web-6b8f7d9c5-zr7tm    0/1     Running   0          8m
```

`Running`, mas `0/1`: os processos estão vivos e nenhum pod está pronto.

## 2. Para onde o Service manda o tráfego?

```bash
kubectl describe svc loja-web | grep -i endpoints
```

```
Endpoints:
```

Vazio. O nginx repassa para o Service, e o Service não tem nenhum pod para atender.

## 3. Por que os pods não estão prontos?

```bash
kubectl describe pod loja-web-6b8f7d9c5-4kq2x | tail -4
```

```
Warning  Unhealthy  ... Readiness probe failed: HTTP probe failed with statuscode: 404
```

```bash
kubectl get deploy loja-web -o jsonpath='{.spec.template.spec.containers[0].readinessProbe}'; echo
```

```
{"httpGet":{"path":"/healthz","port":8080,"scheme":"HTTP"},...}
```

## 4. Teste de dentro do pod

```bash
kubectl exec loja-web-6b8f7d9c5-4kq2x -- wget -qO- http://localhost:8080/health
kubectl exec loja-web-6b8f7d9c5-4kq2x -- wget -qO- http://localhost:8080/healthz
```

```
ok versao=2.6 db=mysql://loja@db.loja.svc:3306/loja
wget: server returned error: HTTP/1.1 404 Not Found
```

A padronização trocou o caminho da probe para `/healthz`, mas a loja só atende `/health`.

## 5. Corrija a probe

```bash
kubectl patch deploy loja-web --type json -p \
  '[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/health"}]'
kubectl rollout status deploy/loja-web
kubectl get pods
kubectl describe svc loja-web | grep -i endpoints
curl -I http://localhost/
```

```
Endpoints:  10.42.0.21:8080,10.42.0.22:8080
HTTP/1.1 200 OK
```

`kubectl rollout undo deploy/loja-web` também funcionaria, porque desfaz a mudança da probe. O que **não** vale é apagar a probe: a loja volta, mas fica desprotegida em todo deploy.

## E a padronização?

O correto a longo prazo é a aplicação atender `/healthz` também, e só depois a probe mudar. A ordem importa: primeiro a aplicação, depois quem a verifica.
