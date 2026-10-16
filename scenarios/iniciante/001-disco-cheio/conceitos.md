# Conceitos para este cenário

## Sistemas de arquivos e partições

No Linux, tudo fica numa única árvore que começa em `/`, mas partes dela podem estar em discos ou partições diferentes. É comum `/var/log` ficar numa partição separada justamente para que logs descontrolados não derrubem o sistema inteiro.

Quando uma partição enche, **qualquer programa que tente escrever nela falha**, mesmo que as outras tenham espaço livre.

## `df` x `du`

| Comando | Responde | Exemplo |
|---|---|---|
| `df -h` | Quanto cada **partição** tem livre? | `df -h /var/log` |
| `du -sh` | Quanto cada **diretório/arquivo** ocupa? | `du -sh /var/log/*` |

Fluxo clássico: `df` para achar a partição cheia, `du` para descer até o culpado.

Dica: `du -sh * | sort -h` ordena do menor para o maior; o culpado aparece no fim.

## Logs e níveis de log

Aplicações registram eventos em arquivos, geralmente em `/var/log`. O **nível de log** define o quanto se registra:

| Nível | Uso |
|---|---|
| `DEBUG` | Tudo, nos mínimos detalhes. Para investigação pontual. |
| `INFO` | Eventos normais de funcionamento. Padrão de produção. |
| `WARNING` | Algo estranho, mas o serviço segue funcionando. |
| `ERROR` | Falhas. |

`DEBUG` em produção pode gerar gigabytes em poucas horas.

## systemd: serviços

| Comando | Para quê |
|---|---|
| `systemctl status app` | Estado do serviço, PID e últimas mensagens |
| `systemctl restart app` | Reinicia (relê a configuração) |
| `journalctl -u app -n 50` | Últimas 50 linhas de log do serviço |
| `systemctl cat app` | Mostra o arquivo de unit (onde está o executável) |

A maioria dos programas só lê a configuração **ao iniciar**. Editar o arquivo sem reiniciar não muda nada.

## `rm` x `truncate`

- `rm arquivo` remove o **nome** do arquivo. Se algum processo ainda estiver com ele aberto, o espaço **não é liberado** até o processo fechar.
- `truncate -s 0 arquivo` zera o **conteúdo**. O espaço volta na hora, mesmo com o arquivo aberto.

Para logs de serviços rodando, prefira `truncate`. E, antes de apagar, pense: alguém vai precisar desse log para entender o incidente?

## logrotate

Ferramenta que "gira" logs periodicamente: renomeia o atual, comprime os antigos e apaga os mais velhos. As regras ficam em `/etc/logrotate.d/`. Um log sem rotação cresce para sempre.
