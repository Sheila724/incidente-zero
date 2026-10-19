# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Reproduza como um cliente

```bash
curl -v --resolve loja.local:443:127.0.0.1 https://loja.local/
```

```
curl: (60) SSL certificate problem: certificate has expired
```

## 2. Veja o certificado em uso

```bash
openssl s_client -connect 127.0.0.1:443 -servername loja.local </dev/null 2>/dev/null \
  | openssl x509 -noout -subject -issuer -dates
```

```
subject=CN = loja.local
issuer=O = Loja Exemplo, CN = CA Interna Loja Exemplo
notBefore=...
notAfter=...            ← ontem
```

Emitido pela CA interna, válido por 90 dias, venceu ontem.

## 3. Ache a renovação automática

```bash
systemctl list-timers --all | grep -i renova      # não aparece
systemctl list-unit-files '*.timer' | grep renova
```

```
renova-certificado.timer   disabled   enabled
```

O timer existe, mas está **desabilitado**: na migração para o template novo, os arquivos foram copiados, mas ninguém rodou `systemctl enable`.

```bash
systemctl cat renova-certificado.timer renova-certificado.service
cat /usr/local/sbin/renova-certificado
```

O script renova quando faltam menos de 30 dias, grava o arquivo... e termina. Ele **não recarrega o nginx**.

## 4. Renove

```bash
sudo systemctl start renova-certificado.service
journalctl -u renova-certificado -n 5 --no-pager
openssl x509 -noout -dates -in /etc/ssl/loja/loja.crt
```

O arquivo agora vale por mais 90 dias.

## 5. A armadilha: o site continua com o certificado vencido

```bash
curl -s -o /dev/null -w '%{http_code}\n' --resolve loja.local:443:127.0.0.1 https://loja.local/
```

Ainda falha. O nginx carregou o certificado antigo na memória:

```bash
sudo nginx -t && sudo systemctl reload nginx
curl -i --resolve loja.local:443:127.0.0.1 https://loja.local/
```

**200**. Incidente mitigado.

## 6. Conserte a renovação de ponta a ponta

Habilite o timer:

```bash
sudo systemctl enable --now renova-certificado.timer
systemctl list-timers renova-certificado.timer
```

E faça a renovação aplicar o certificado novo. Duas opções:

**No script** (acrescente no fim de `/usr/local/sbin/renova-certificado`):

```bash
nginx -t && systemctl reload nginx
```

**Ou no serviço**, sem mexer no script:

```bash
sudo systemctl edit renova-certificado.service
```

```ini
[Service]
ExecStartPost=/usr/bin/systemctl reload nginx
```

Agora, da próxima vez, o timer dispara, o script emite e o nginx passa a usar o certificado novo, sem ninguém acordar às 8h.
