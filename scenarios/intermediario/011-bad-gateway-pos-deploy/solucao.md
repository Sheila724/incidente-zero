# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Confirme o sintoma

```bash
curl -i http://localhost/
curl -i http://localhost/api/produtos
```

Os dois respondem **502 Bad Gateway**, com a página de erro do próprio nginx. O nginx está de pé; o problema é atrás dele.

## 2. Leia o log do nginx

```bash
sudo tail -n 5 /var/log/nginx/error.log
```

```
connect() failed (111: Connection refused) while connecting to upstream ...
upstream: "http://127.0.0.1:8080/"
```

O nginx tenta falar com a aplicação na porta **8080** e ninguém atende.

## 3. Onde a aplicação está?

```bash
systemctl status app --no-pager
sudo ss -ltnp | grep python
cat /etc/app/app.env
```

```
LISTEN  127.0.0.1:8081  users:(("python3",...))
PORT=8081
```

A aplicação está saudável, na **8081**. O changelog explica:

```bash
cat /opt/app/CHANGELOG.md
```

> A aplicação passa a escutar na porta 8081. A 8080 fica reservada para o agente de métricas (DEP-87). Ajustar o proxy reverso antes do deploy (DEP-88).

O ajuste do proxy ficou para trás. Voltar a aplicação para a 8080 resolveria hoje e quebraria na próxima semana, quando o agente de métricas chegar.

## 4. Encontre todas as referências

```bash
grep -rn '127.0.0.1:80' /etc/nginx/
```

```
/etc/nginx/conf.d/upstream-loja.conf:3:    server 127.0.0.1:8080;
/etc/nginx/sites-available/loja:16:        proxy_pass http://127.0.0.1:8080;
```

**Dois lugares.** A `location /api/` não usa o upstream: tem o endereço escrito direto.

## 5. A armadilha: corrigir só o upstream

Quem corrige só o primeiro arquivo vê a home voltar com **200** e encerra o incidente, mas o app mobile continua quebrado:

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost/api/produtos    # 502
```

## 6. Corrija os dois

```bash
sudo nano /etc/nginx/conf.d/upstream-loja.conf
```

```nginx
upstream loja_app {
    server 127.0.0.1:8081;
    keepalive 16;
}
```

```bash
sudo nano /etc/nginx/sites-available/loja
```

```nginx
    location /api/ {
        proxy_pass http://loja_app;
        proxy_read_timeout 5s;
    }
```

Usar o upstream também na `/api/` deixa o endereço num lugar só.

## 7. Teste e aplique

```bash
sudo nginx -t && sudo systemctl reload nginx
curl -i http://localhost/
curl -s http://localhost/api/produtos
```

Site **200** e API respondendo `"versao": "2.4.0"`. Incidente mitigado, com o deploy mantido.
