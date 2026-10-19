# Conceitos para este cenário

## O que é um certificado TLS

Um certificado liga um **nome** (`loja.local`) a uma **chave pública** e é assinado por uma **autoridade certificadora (CA)**. O cliente confere três coisas antes de confiar:

| Verificação | Erro típico quando falha |
|---|---|
| O certificado foi assinado por uma CA em que eu confio? | `unable to get local issuer certificate` |
| O nome do site está no certificado (SAN)? | `no alternative certificate subject name matches` |
| A data de hoje está entre `notBefore` e `notAfter`? | `certificate has expired` |

## CA interna

Empresas costumam ter uma CA própria para sistemas internos. O certificado dela é instalado como confiável nos servidores e clientes (em Ubuntu: `/usr/local/share/ca-certificates/` + `update-ca-certificates`).

## Inspecionando certificados com openssl

Um arquivo no disco:

```bash
openssl x509 -noout -subject -issuer -dates -in /caminho/cert.crt
openssl x509 -checkend 2592000 -noout -in cert.crt   # vence nos próximos 30 dias?
```

O que o **servidor está entregando agora** (pode ser diferente do arquivo!):

```bash
openssl s_client -connect 127.0.0.1:443 -servername loja.local </dev/null \
  | openssl x509 -noout -dates
```

O `-servername` envia o SNI: com vários sites na mesma porta, é ele que escolhe o certificado.

## Arquivo no disco × certificado em uso

O nginx lê o certificado **quando inicia ou recarrega** e o mantém em memória. Trocar o arquivo não muda nada até um `reload`:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

O `reload` não derruba conexões: os processos novos assumem e os antigos terminam o que estavam fazendo. Por isso é preferível ao `restart`.

## Timers do systemd

Um `.timer` dispara um `.service` com o mesmo nome no horário configurado (`OnCalendar=daily`). Para ver todos:

```bash
systemctl list-timers --all            # timers carregados, com próxima execução
systemctl list-unit-files '*.timer'    # inclusive os desabilitados
```

Um timer precisa estar **enabled** (sobe no boot) e **active** (rodando agora).

## Renovação de ponta a ponta

Renovar tem três partes, e todas precisam funcionar:

1. **Disparar** (timer ou cron);
2. **Emitir** o certificado novo;
3. **Aplicar** (recarregar quem usa o certificado: nginx, Apache, HAProxy...).

O certbot chama a parte 3 de *deploy hook*. Esquecer dela é um dos motivos mais comuns de "o certificado renovou, mas o site continua com o vencido".
