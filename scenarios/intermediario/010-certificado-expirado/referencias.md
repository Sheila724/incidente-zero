# Para estudar mais

## TLS e certificados

- [How HTTPS works](https://howhttps.works/): quadrinho sobre TLS, certificados e CAs
- [openssl-x509](https://docs.openssl.org/3.0/man1/openssl-x509/): inspecionar e conferir certificados
- [openssl-s_client](https://docs.openssl.org/3.0/man1/openssl-s_client/): testar o que um servidor entrega

## nginx

- [Configurando servidores HTTPS](https://nginx.org/en/docs/http/configuring_https_servers.html)
- [Controlando o nginx](https://nginx.org/en/docs/control.html): `reload` sem derrubar conexões

## systemd

- [systemd.timer](https://www.freedesktop.org/software/systemd/man/latest/systemd.timer.html): `OnCalendar`, `Persistent`
- [Hooks do certbot](https://eff-certbot.readthedocs.io/en/stable/using.html#renewing-certificates): `--deploy-hook` na renovação

## Para praticar

- Gere uma CA e um certificado com `openssl` numa VM e configure o nginx com ele.
- Monitore a validade com um item do Zabbix (`web.certificate.get`) ou com `openssl -checkend` num cron.
- Próximo cenário relacionado: **011 — 502 Bad Gateway após deploy**.
