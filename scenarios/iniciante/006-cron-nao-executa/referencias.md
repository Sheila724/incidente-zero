# Para estudar mais

## cron

- [cron(8)](https://manpages.ubuntu.com/manpages/noble/man8/cron.8.html): veja a seção sobre os nomes de arquivo em `/etc/cron.d`
- [crontab(5)](https://manpages.ubuntu.com/manpages/noble/man5/crontab.5.html): formato da linha e o significado do `%`
- [crontab.guru](https://crontab.guru): tradutor de expressões do cron

## Alternativa moderna

- [systemd.timer](https://www.freedesktop.org/software/systemd/man/latest/systemd.timer.html): timers do systemd, com log no journal e `systemctl list-timers`

## Backup

- [tar](https://man7.org/linux/man-pages/man1/tar.1.html): `-czf` para criar, `-tzf` para listar

## Para praticar

- Crie um agendamento a cada minuto que escreva a data num arquivo e acompanhe pelo journal.
- Coloque um `%` sem escape de propósito e veja o que aparece no log.
- Próximo cenário relacionado: **003 — Permissão negada** (outro exemplo de timer do systemd).
