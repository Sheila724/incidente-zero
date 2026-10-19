# Para estudar mais

## nginx como proxy reverso

- [ngx_http_proxy_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html): `proxy_pass`, timeouts e cabeçalhos
- [ngx_http_upstream_module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html): `upstream`, `keepalive` e balanceamento
- [Controlando o nginx](https://nginx.org/en/docs/control.html): `reload` sem derrubar conexões

## HTTP

- [Códigos de status 5xx (MDN)](https://developer.mozilla.org/pt-BR/docs/Web/HTTP/Status#respostas_de_erro_do_servidor): 502, 503 e 504

## Rede

- [ss](https://man7.org/linux/man-pages/man8/ss.8.html): portas em escuta e conexões

## Para praticar

- Suba duas instâncias de uma aplicação em portas diferentes e balanceie com um `upstream`.
- Compare um 502 (aplicação parada) com um 504 (aplicação lenta) mexendo no `proxy_read_timeout`.
- Cenário relacionado: **005 — Porta 80 ocupada**.
