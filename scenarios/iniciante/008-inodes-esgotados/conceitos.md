# Conceitos para este cenário

## O que é um inode

Todo arquivo ou pasta no Linux tem um **inode**: uma ficha com dono, permissões, datas e onde os dados estão no disco. O nome do arquivo fica na pasta, apontando para o inode.

A quantidade de inodes de um sistema de arquivos é **definida na criação** e não aumenta depois. Então dá para acabar de dois jeitos:

| Acabou | Sintoma | Como ver |
|---|---|---|
| Espaço | Arquivos grandes enchem o disco | `df -h` |
| Inodes | Muitos arquivos pequenos, mesmo com espaço sobrando | `df -i` |

Nos dois casos o erro é o mesmo: `No space left on device` (`ENOSPC`, erro 28).

## `df -i`

```bash
df -i /var/lib/app
# Filesystem  Inodes IUsed IFree IUse% Mounted on
# tmpfs         3000  3000     0  100% /var/lib/app
```

## Achando onde estão os arquivos

```bash
sudo du --inodes -x /var/lib/app | sort -n | tail     # inodes por pasta
sudo find /var/lib/app -xdev -type f | wc -l           # total de arquivos
ls -lt /var/lib/app/sessions | tail -3                 # os mais antigos
```

`-x` / `-xdev` evita entrar em outros sistemas de arquivos.

## Apagando milhares de arquivos

```bash
rm /pasta/*
# bash: /usr/bin/rm: Argument list too long
```

Com muitos arquivos, o `*` gera uma linha de comando grande demais. Use o `find`, que apaga um por um:

```bash
sudo find /var/lib/app/sessions -type f -mmin +60 -delete
```

| Opção | Significado |
|---|---|
| `-type f` | Só arquivos |
| `-mmin +60` | Modificados há mais de 60 minutos |
| `-mtime +7` | Modificados há mais de 7 dias |
| `-delete` | Apaga o que encontrou |

Rode **sem** o `-delete` primeiro para ver o que vai ser apagado.

## Sessões

Uma aplicação web guarda a sessão de cada visitante (carrinho, login) em algum lugar, aqui em um arquivo por sessão. Sessões antigas precisam ser limpas periodicamente; as recentes são de clientes navegando **agora**, e apagá-las desloga essas pessoas.

## Timers do systemd

```bash
systemctl list-timers --all             # agendados, inclusive os desativados
sudo systemctl enable --now algo.timer  # ativa agora e no boot
```
