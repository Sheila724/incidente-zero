# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Reproduza o sintoma

```bash
curl -I http://localhost/
curl -I http://localhost/produtos/
curl -I http://localhost/img/logo.svg
```

A home dá **403**, mas produtos e o logo dão **404**, apesar de existirem:

```bash
ls /var/www/loja/produtos/      # permissão negada até para o aluno, sem sudo
sudo ls /var/www/loja/produtos/ # o arquivo está lá
```

## 2. Leia o log de erros

```bash
sudo tail -n 20 /var/log/nginx/error.log
```

```
stat() "/var/www/loja/img/logo.svg" failed (13: Permission denied)
```

Erro 13: permissão negada. O problema não é o nginx nem os arquivos, é o **acesso**.

## 3. Descubra quem precisa de acesso

```bash
ps -o user,pid,cmd -C nginx
```

Os workers rodam como `www-data`.

## 4. Inspecione o caminho

```bash
namei -l /var/www/loja/index.html
```

```
drwxr-xr-x root   root   /
drwxr-xr-x root   root   var
drwxr-xr-x root   root   www
drwxrwx--- deploy deploy loja
```

Dono e grupo são `deploy`, e **outros** (onde cai o `www-data`) não têm nada. O nginx não consegue nem entrar na pasta.

## 5. Mitigação rápida

```bash
sudo chmod -R o+rX /var/www/loja
curl -I http://localhost/
```

**200**. Mas espere uns 2 minutos e teste de novo: o 403 volta.

## 6. Ache quem desfaz a correção

```bash
systemctl list-timers --all
```

Aparece o `hardening-web.timer`, rodando a cada 2 minutos:

```bash
systemctl cat hardening-web.timer hardening-web.service
cat /usr/local/bin/hardening-web.sh
```

O script tira **toda** permissão de "outros" para atender a auditoria SEC-142. A intenção é boa, mas o nginx também está em "outros".

## 7. Corrija a causa: atender a auditoria sem bloquear o nginx

A solução certa não é liberar para todos, e sim colocar o nginx no **grupo** dos arquivos:

```bash
sudo nano /usr/local/bin/hardening-web.sh
```

```bash
chown -R deploy:www-data /var/www/loja
chmod -R u=rwX,g=rX,o= /var/www/loja
```

Aplique na hora e valide:

```bash
sudo systemctl start hardening-web.service
namei -l /var/www/loja/index.html
curl -I http://localhost/
curl -I http://localhost/img/logo.svg
```

Resultado: `deploy` edita, `www-data` só lê e **ninguém mais** acessa. A auditoria continua atendida e o site no ar.

## Alternativa: desativar o timer

```bash
sudo systemctl disable --now hardening-web.timer
sudo chmod -R o+rX /var/www/loja
```

Resolve o incidente, mas deixa a auditoria pendente. Se fizer isso, registre e combine com o time de segurança.
