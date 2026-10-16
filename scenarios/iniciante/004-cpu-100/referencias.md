# Para estudar mais

## Processos e CPU

- [top](https://man7.org/linux/man-pages/man1/top.1.html): monitor de processos em tempo real
- [ps](https://man7.org/linux/man-pages/man1/ps.1.html): seleção e formatação de processos
- [kill](https://man7.org/linux/man-pages/man1/kill.1.html) e [signal(7)](https://man7.org/linux/man-pages/man7/signal.7.html): sinais e seus efeitos
- [proc(5)](https://man7.org/linux/man-pages/man5/proc.5.html): o que existe em `/proc/<PID>/`

## systemd

- [systemd.service](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html): `Restart=` e outras opções
- [systemd.resource-control](https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html): `CPUQuota=`, `MemoryMax=`

## Para praticar

- Rode `yes > /dev/null &` na sua máquina, encontre o processo no `top` e encerre com `kill`.
- Crie um serviço de teste com `Restart=always` e veja o que acontece ao matar o processo.
- Próximo cenário relacionado: **005 — Porta já em uso**.
