# Para estudar mais

## Resolução de nomes no Linux

- [nsswitch.conf](https://man7.org/linux/man-pages/man5/nsswitch.conf.5.html): ordem das fontes de nomes
- [hosts](https://man7.org/linux/man-pages/man5/hosts.5.html): formato do `/etc/hosts`
- [resolv.conf](https://man7.org/linux/man-pages/man5/resolv.conf.5.html): servidores DNS, `search` e opções
- [getent](https://man7.org/linux/man-pages/man1/getent.1.html): consultar as bases do sistema como uma aplicação

## DNS

- [dig](https://bind9.readthedocs.io/en/latest/manpages.html): consultas, `+short`, `+trace` e `@servidor`
- [How DNS works](https://howdns.works/): quadrinho que explica o caminho de uma consulta

## Para praticar

- Coloque um nome no seu `/etc/hosts` e compare `dig` e `getent` para ele.
- Troque a ordem de `hosts:` no `nsswitch.conf` de uma VM de teste e veja o que muda.
- Próximo cenário relacionado: **011 — 502 Bad Gateway após deploy**.
