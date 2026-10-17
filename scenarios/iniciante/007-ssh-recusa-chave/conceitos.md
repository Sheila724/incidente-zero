# Conceitos para este cenário

## Autenticação por chave, em resumo

1. O cliente tem a **chave privada** (`deploy_key`).
2. O servidor tem a **chave pública** correspondente no `~/.ssh/authorized_keys` do usuário.
3. No login, o cliente prova que tem a privada, sem enviá-la.

## O cliente não diz o motivo

`Permission denied (publickey)` é tudo o que o cliente recebe, **de propósito**: dar detalhes ajudaria um atacante. O motivo real fica no log do servidor:

```bash
journalctl -u ssh -n 30
```

Do lado do cliente, `ssh -v` mostra quais chaves foram oferecidas e o que o servidor respondeu.

## StrictModes: as permissões que o sshd exige

Por padrão (`StrictModes yes`), o sshd **recusa** o `authorized_keys` se alguém além do dono puder alterá-lo, nem que seja indiretamente:

| Caminho | Permissão recomendada | Por quê |
|---|---|---|
| `/home/deploy` | `755` ou `750` | Ninguém além do dono pode gravar |
| `~/.ssh` | `700` | Só o dono entra |
| `~/.ssh/authorized_keys` | `600` | Só o dono lê e grava |

Dono de tudo: o próprio usuário (`chown deploy:deploy`).

Se a home fosse gravável por todos, qualquer usuário do servidor poderia trocar a pasta `.ssh` e colocar a própria chave lá. Por isso o sshd prefere recusar.

## Lendo permissões do caminho inteiro

```bash
namei -l /home/deploy/.ssh/authorized_keys
```

## Por que `chmod -R 777` é sempre um erro

`777` dá leitura, escrita e execução para **todos** os usuários do sistema. Além de quebrar o SSH, deixa qualquer processo invadido alterar scripts, configurações e código. Quase sempre o problema real era de **dono** (`chown`) ou de **grupo**.

## O que *não* fazer para "resolver"

- `StrictModes no` no `sshd_config`: desliga a proteção para todo mundo.
- `PasswordAuthentication yes`: abre o servidor para ataques de força bruta.
