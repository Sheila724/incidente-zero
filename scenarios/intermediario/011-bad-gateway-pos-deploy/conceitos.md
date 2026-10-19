# Conceitos para este cenário

## Proxy reverso

```
cliente ──▶ nginx :80 ──▶ aplicação 127.0.0.1:PORTA
```

O nginx recebe a requisição e a repassa (`proxy_pass`) para a aplicação, que só escuta em `127.0.0.1`. Ele é o único ponto exposto, e cuida de TLS, cache, timeouts e de distribuir carga entre várias instâncias.

## Os códigos 5xx de um proxy

| Código | O que o proxy está dizendo |
|---|---|
| **502 Bad Gateway** | Não consegui uma resposta válida do backend: conexão recusada, conexão fechada no meio ou resposta inválida. |
| **503 Service Unavailable** | Não há backend disponível, ou o próprio backend disse que está indisponível. |
| **504 Gateway Timeout** | O backend aceitou a conexão, mas demorou mais que o `proxy_read_timeout`. |

O `error.log` do nginx diz o motivo e o endereço exato:

```
connect() failed (111: Connection refused) while connecting to upstream,
upstream: "http://127.0.0.1:8080/"
```

`111: Connection refused` = nada escutando ali.

## `upstream`: o backend com nome

```nginx
upstream loja_app {
    server 127.0.0.1:8080;
    keepalive 16;
}

location / {
    proxy_pass http://loja_app;
}
```

O bloco `upstream` dá um nome ao backend. Todas as rotas que usam `http://loja_app` mudam juntas quando o endereço muda. Já um `proxy_pass http://127.0.0.1:8080` escrito direto numa `location` é uma segunda fonte de verdade, fácil de esquecer.

## Onde cada processo escuta

```bash
sudo ss -ltnp
```

| Coluna | Significado |
|---|---|
| `Local Address:Port` | Endereço e porta em que o processo aceita conexões |
| `Process` | Nome e PID (precisa de `sudo` para ver processos de outros usuários) |

`127.0.0.1:8081` só aceita conexões da própria máquina; `0.0.0.0:80`, de qualquer interface.

## Encontrando todas as referências

Antes de mudar um endereço, procure todos os lugares que o usam:

```bash
grep -rn '127.0.0.1:8080' /etc/nginx/
sudo nginx -T | grep -n proxy_pass      # a configuração completa, já com os includes
```

## Mudar o nginx com segurança

```bash
sudo nginx -t && sudo systemctl reload nginx
```

O `-t` valida a sintaxe; o `reload` aplica sem derrubar as conexões abertas. Se o teste falhar, o `&&` impede o reload e a configuração antiga continua no ar.
