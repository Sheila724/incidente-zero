# Para estudar mais

## Inodes e sistemas de arquivos

- [inode(7)](https://man7.org/linux/man-pages/man7/inode.7.html): o que há dentro de um inode
- [df](https://man7.org/linux/man-pages/man1/df.1.html): a opção `-i`
- [du](https://man7.org/linux/man-pages/man1/du.1.html): a opção `--inodes`
- [stat](https://man7.org/linux/man-pages/man1/stat.1.html): veja o número do inode de um arquivo

## find

- [find](https://man7.org/linux/man-pages/man1/find.1.html): `-mmin`, `-mtime`, `-delete`, `-xdev`

## systemd

- [systemd.timer](https://www.freedesktop.org/software/systemd/man/latest/systemd.timer.html)

## Para praticar

- Rode `df -i` na sua máquina. Qual partição tem a maior porcentagem de inodes usados?
- Crie uma pasta com 10 mil arquivos (`touch arquivo{1..10000}`) e tente `rm *`.
- Cenário relacionado: **001 — Disco cheio em /var/log** (o caso do espaço).
