# Para estudar mais

## nginx

- [Guia para iniciantes](https://nginx.org/en/docs/beginners_guide.html): estrutura da configuração, recarregar e testar
- [Como o nginx processa uma requisição](https://nginx.org/en/docs/http/request_processing.html): `server_name` e `default_server`
- [Server names](https://nginx.org/en/docs/http/server_names.html): curingas, regex e nome padrão

## systemd

- [systemctl](https://www.freedesktop.org/software/systemd/man/latest/systemctl.html): `status`, `restart`, `reload`
- [journalctl](https://www.freedesktop.org/software/systemd/man/latest/journalctl.html): filtros por serviço e por tempo

## Rede

- [ss](https://man7.org/linux/man-pages/man8/ss.8.html): quais processos estão escutando em quais portas

## Para praticar

- Crie um segundo site no seu nginx local e teste com `curl -H "Host: ..."`.
- Quebre a configuração de propósito e compare o que acontece com `reload` e com `restart`.
- Próximo cenário relacionado: **005 — Porta já em uso**.
