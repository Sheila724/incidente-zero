# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Delimite o problema

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost/
curl -i http://localhost/checkout
```

A home responde **200**, o checkout responde **502** com `gateway de pagamento indisponível`. O problema está numa dependência do checkout.

## 2. Leia o log da aplicação

```bash
journalctl -u app -n 20 --no-pager
```

```
pagamentos: api.pagamentos.interno -> 10.20.0.9 (resolvido na inicialização)
ERRO checkout: falha ao chamar api.pagamentos.interno (10.20.0.9:9000): timed out
```

A aplicação chama a API em **10.20.0.9**, o servidor antigo, que foi desligado na migração. Por isso o `timed out`.

## 3. Compare o DNS com o sistema

```bash
dig +short api.pagamentos.interno
getent hosts api.pagamentos.interno
grep '^hosts:' /etc/nsswitch.conf
```

```
10.20.0.15                                  ← DNS: já está certo
10.20.0.9       api.pagamentos.interno      ← sistema: ainda no IP antigo
hosts:          files dns                   ← /etc/hosts vem antes do DNS
```

O time de pagamentos tinha razão: o DNS está certo. Mas este servidor nem pergunta ao DNS, porque acha o nome antes no `/etc/hosts`.

## 4. Ache a linha esquecida

```bash
grep -n pagamentos /etc/hosts
```

```
# workaround temporário (INC-2291): DNS interno instável, fixando o IP da API
10.20.0.9    api.pagamentos.interno
```

Um "workaround temporário" de outro incidente, que nunca foi desfeito.

## 5. Remova a entrada

```bash
sudo nano /etc/hosts
```

Apague as duas linhas (o comentário e a entrada) e salve. Confira:

```bash
getent hosts api.pagamentos.interno     # agora: 10.20.0.15
```

> Num lab, o `/etc/hosts` é montado pelo Docker, então `sed -i` falha com *Device or resource busy*: ele tenta substituir o arquivo, e um ponto de montagem não pode ser substituído. Editores como o `nano` gravam no próprio arquivo e funcionam.

## 6. A armadilha: o checkout continua com 502

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost/checkout    # ainda 502
curl -s http://127.0.0.1:8080/health
```

```json
{"status": "ok", "pagamentos": "api.pagamentos.interno", "pagamentos_ip": "10.20.0.9"}
```

O sistema já resolve certo, mas a aplicação guardou o IP quando iniciou. Reinicie:

```bash
sudo systemctl restart app
curl -i http://localhost/checkout
```

**200** com `pedido confirmado`. Incidente mitigado.

## Por que não colocar o IP novo no /etc/hosts?

Funcionaria hoje, mas repetiria o erro: na próxima migração, este servidor quebraria de novo. O nome precisa vir do DNS, que é a fonte que o time de pagamentos atualiza.
