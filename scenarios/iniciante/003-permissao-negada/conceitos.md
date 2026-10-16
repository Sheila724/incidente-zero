# Conceitos para este cenário

## Lendo permissões

```
drwxr-x--- 4 deploy deploy 4096 out 16 10:30 /var/www/loja
│└┬┘└┬┘└┬┘   └─dono─┘ └grupo┘
│ │  │  └── outros (todo o resto)
│ │  └───── grupo
│ └──────── dono
└────────── d = diretório
```

| Letra | Em arquivo | Em pasta |
|---|---|---|
| `r` | ler o conteúdo | listar o que tem dentro |
| `w` | alterar | criar e apagar itens |
| `x` | executar | **atravessar** (entrar e acessar o que tem dentro) |

Para ler `/var/www/loja/index.html`, um usuário precisa de `x` em **todas** as pastas do caminho e `r` no arquivo.

## Quem é o nginx?

O processo principal do nginx roda como root, mas quem atende as requisições são os *workers*, que rodam como **`www-data`**:

```bash
ps -o user,pid,cmd -C nginx
```

Se o dono dos arquivos é `deploy` e o grupo é `deploy`, para o `www-data` valem as permissões de **outros**.

## `namei -l`: permissões do caminho inteiro

```bash
namei -l /var/www/loja/img/logo.svg
```

Mostra dono, grupo e permissões de cada pasta do caminho. Ótimo para achar onde o acesso trava.

## 403 x 404 em problemas de permissão

Quando o nginx não consegue nem verificar se um arquivo existe (falta `x` na pasta), o resultado depende da configuração: pode vir **403** ou até **404**. Um 404 nem sempre significa que o arquivo não existe. Na dúvida, o `error.log` tira a dúvida:

```
stat() "/var/www/loja/img/logo.svg" failed (13: Permission denied)
```

O erro **13** é `EACCES`: permissão negada.

## `chmod` simbólico e o `X` maiúsculo

```bash
chmod -R u=rwX,g=rX,o= /var/www/loja
```

- `u=rwX`: dono lê e escreve.
- `g=rX`: grupo lê.
- `o=`: outros não têm nada.
- **`X` maiúsculo** dá `x` só para pastas (e arquivos que já eram executáveis). Assim pastas ficam atravessáveis sem tornar os HTML executáveis.

## Por que não `chmod 777`?

`777` deixa **qualquer** usuário ou processo do servidor alterar o site. Se alguma aplicação for invadida, o atacante troca suas páginas. Dê o mínimo necessário: o nginx só precisa **ler**.

## Timers do systemd

Além do cron, o systemd agenda tarefas com *timers*: um `.timer` dispara um `.service`.

```bash
systemctl list-timers --all         # o que está agendado e quando roda
systemctl cat hardening-web.timer   # ver a definição
systemctl disable --now algo.timer  # desativar
```
