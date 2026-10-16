# Para estudar mais

## Permissões

- [chmod](https://man7.org/linux/man-pages/man1/chmod.1.html): modo simbólico, numérico e o `X` maiúsculo
- [chown](https://man7.org/linux/man-pages/man1/chown.1.html): dono e grupo
- [namei](https://man7.org/linux/man-pages/man1/namei.1.html): permissões de cada parte de um caminho
- [path_resolution](https://man7.org/linux/man-pages/man7/path_resolution.7.html): como o kernel verifica permissões ao percorrer um caminho

## systemd

- [systemd.timer](https://www.freedesktop.org/software/systemd/man/latest/systemd.timer.html): agendamento com timers
- [systemctl](https://www.freedesktop.org/software/systemd/man/latest/systemctl.html): `list-timers`, `cat`, `disable --now`

## nginx

- [Guia para iniciantes](https://nginx.org/en/docs/beginners_guide.html): estrutura e logs

## Para praticar

- Rode `namei -l` em arquivos do seu sistema, como `/etc/shadow`. Quem consegue ler?
- Crie uma pasta com `chmod 711` e outra com `chmod 744`. Em qual delas outro usuário consegue abrir um arquivo cujo nome ele já sabe?
- Próximo cenário relacionado: **007 — SSH recusando a chave** (permissões do `authorized_keys`).
