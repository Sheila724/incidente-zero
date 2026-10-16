# Para estudar mais

## Comandos (man pages)

- [df](https://man7.org/linux/man-pages/man1/df.1.html): espaço livre por sistema de arquivos
- [du](https://man7.org/linux/man-pages/man1/du.1.html): espaço ocupado por arquivos e diretórios
- [truncate](https://man7.org/linux/man-pages/man1/truncate.1.html): altera o tamanho de um arquivo
- [systemctl](https://www.freedesktop.org/software/systemd/man/latest/systemctl.html): controle de serviços do systemd
- [journalctl](https://www.freedesktop.org/software/systemd/man/latest/journalctl.html): leitura do journal do systemd
- [logrotate](https://github.com/logrotate/logrotate): projeto e documentação

## SRE

- [Effective Troubleshooting](https://sre.google/sre-book/effective-troubleshooting/): capítulo do livro de SRE do Google sobre método de diagnóstico
- [Postmortem Culture](https://sre.google/sre-book/postmortem-culture/): como escrever postmortems sem culpados

## Para praticar

- Rode `df -h` e `sudo du -sh /var/* | sort -h` na sua própria máquina. O que mais ocupa espaço?
- Leia as regras em `/etc/logrotate.d/` de um servidor Linux e tente explicar cada linha.
- Próximo cenário relacionado: **008 — "Disco cheio" com espaço sobrando** (inodes esgotados).
