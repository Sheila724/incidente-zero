# Conceitos para este cenário

## Siga a requisição

Num incidente com várias peças, desenhe o caminho da requisição e teste cada salto:

```
cliente ──▶ nginx (:80) ──▶ app (127.0.0.1:8080) ──▶ API de pagamentos (api.pagamentos.interno:9000)
```

Um **502 Bad Gateway** quer dizer que o nginx está de pé, mas recebeu uma resposta ruim (ou nenhuma) de quem está atrás dele. Aqui, a aplicação também está de pé: quem devolve o erro é ela, porque a dependência dela falhou.

## Como o Linux transforma um nome em IP

Quando um programa pede o IP de um nome (função `getaddrinfo`), o sistema consulta as fontes na ordem definida em `/etc/nsswitch.conf`:

```
hosts:  files dns
```

| Fonte | Onde fica | Observação |
|---|---|---|
| `files` | `/etc/hosts` | Consultado **primeiro**. Uma linha aqui ganha do DNS. |
| `dns` | servidores em `/etc/resolv.conf` | Só é consultado se o nome não estiver no `/etc/hosts`. |

## `dig` × `getent`

| Comando | O que mostra |
|---|---|
| `dig nome` | A resposta do **servidor DNS**. Ignora o `/etc/hosts`. |
| `getent hosts nome` | A resposta do **sistema**, seguindo o `nsswitch.conf`, igual à aplicação. |

Quando os dois discordam, a diferença está no caminho entre eles: quase sempre, uma linha esquecida no `/etc/hosts`.

## `/etc/hosts`: atalho perigoso

Fixar um IP no `/etc/hosts` resolve um problema na hora (DNS instável, teste de migração), mas cria outro: aquele servidor para de acompanhar as mudanças do DNS. Meses depois, ninguém lembra da linha, e a próxima migração quebra só nele.

## Cache de DNS na aplicação

Corrigir o sistema nem sempre basta. Muitas aplicações guardam o IP:

- resolvem o nome **uma vez**, na inicialização (como esta);
- usam um pool de conexões que continua aberto para o IP antigo;
- respeitam um TTL longo, ou ignoram o TTL (a JVM, por padrão, pode guardar para sempre).

Depois de corrigir a resolução, confira o que o **processo em execução** está usando: logs, endpoint de health ou `ss -tnp`.

## Timeout × conexão recusada

| Erro | O que costuma significar |
|---|---|
| `Connection refused` | O IP respondeu, mas nada escuta naquela porta. |
| `timed out` | Nada respondeu: IP inexistente, servidor desligado ou firewall descartando. |

Um timeout para um IP que "deveria" funcionar é um bom sinal de que você está falando com o endereço errado.
