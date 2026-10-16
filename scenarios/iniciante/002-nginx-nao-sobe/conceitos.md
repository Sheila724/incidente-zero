# Conceitos para este cenário

## "Fora do ar" x "respondendo com erro"

| O que o `curl` mostra | O que significa |
|---|---|
| `Connection refused` | Nada escutando na porta. O serviço está parado ou caiu. |
| `HTTP 502`, `503` | O servidor web está de pé, mas algo atrás dele falhou. |
| `HTTP 404`, `403` | O servidor responde, mas não acha ou não pode servir o conteúdo. |

Saber ler isso economiza muito tempo: cada um aponta para um lugar diferente.

## Como o nginx organiza a configuração (Debian/Ubuntu)

| Caminho | Para quê |
|---|---|
| `/etc/nginx/nginx.conf` | Configuração principal; inclui os sites |
| `/etc/nginx/sites-available/` | Todos os sites configurados |
| `/etc/nginx/sites-enabled/` | Links para os sites **ativos** |

Ativar um site = criar um link: `ln -s ../sites-available/x /etc/nginx/sites-enabled/x`.
Desativar = remover o link. O arquivo original continua em `sites-available`.

## Blocos `server` e `default_server`

Cada site é um bloco `server`. O nginx escolhe qual bloco atende a requisição pelo cabeçalho `Host`, comparando com o `server_name`.

Se nenhum `server_name` bate (por exemplo, acesso pelo IP), vale o `default_server` daquela porta. **Só pode existir um `default_server` por porta.**

Para testar um site específico sem DNS, envie o `Host` na mão:

```bash
curl -H "Host: promo.loja.local" http://localhost
```

## `nginx -t`: teste antes de aplicar

`nginx -t` valida a sintaxe de **toda** a configuração e aponta arquivo e linha do erro. Ele para no primeiro erro, então rode de novo depois de cada correção.

Regra de ouro: **nunca reinicie o nginx sem rodar `nginx -t` antes.**

## `restart` x `reload`

| Comando | Comportamento | Se a config estiver errada |
|---|---|---|
| `systemctl restart nginx` | Para e sobe de novo | Para e **não volta**: site fora do ar |
| `systemctl reload nginx` | Relê a config sem derrubar | Mantém a config antiga rodando |

Para mudanças de configuração, `reload` é mais seguro.

## Onde ver o motivo de uma falha

```bash
systemctl status nginx            # estado e últimas mensagens
journalctl -u nginx -n 30         # log do serviço no systemd
sudo tail /var/log/nginx/error.log
```
