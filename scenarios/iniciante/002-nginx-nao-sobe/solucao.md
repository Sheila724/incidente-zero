# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Reproduza o sintoma

```bash
curl -i http://localhost
```

`Connection refused`: não há erro HTTP porque **ninguém está escutando** na porta 80. Confirme:

```bash
sudo ss -ltnp | grep :80
```

Nenhuma linha. O nginx não está rodando.

## 2. Veja por que o nginx não sobe

```bash
systemctl status nginx --no-pager
journalctl -u nginx -n 20 --no-pager
```

O serviço está `failed`, e o log mostra que o teste de configuração falhou ao iniciar.

## 3. Teste a configuração

```bash
sudo nginx -t
```

```
[emerg] invalid parameter "server_name" in /etc/nginx/sites-enabled/promo:4
```

O nginx aponta a linha 4, mas o erro está na linha **anterior**: sem o `;`, ele leu `listen 80 default_server server_name ...` como uma instrução só. Erros de sintaxe costumam ser reportados onde o nginx *percebeu* o problema, que pode ser depois de onde ele está.

## 4. Corrija o primeiro erro

```bash
sudo nano /etc/nginx/sites-available/promo
```

Adicione o `;` no fim da linha do `listen` e teste de novo:

```bash
sudo nginx -t
```

```
[emerg] a duplicate default server for 0.0.0.0:80 in /etc/nginx/sites-enabled/promo:3
```

## 5. Corrija o segundo erro

O site da loja já é o `default_server`, e deve continuar sendo: quem acessar pelo IP ou por um nome desconhecido precisa cair na loja, não na promoção. No arquivo da promo, deixe:

```nginx
listen 80;
```

```bash
sudo nginx -t
```

`test is successful`.

## 6. Suba e valide

```bash
sudo systemctl start nginx
curl http://localhost
curl -H "Host: promo.loja.local" http://localhost
```

Loja de volta, e promoção no ar no seu próprio nome.

## Alternativa: rollback

Sob pressão, também é válido **desfazer a mudança primeiro** e corrigir com calma depois:

```bash
sudo rm /etc/nginx/sites-enabled/promo
sudo nginx -t && sudo systemctl start nginx
```

A loja volta em segundos e a promoção sai do ar, mas isso é melhor do que tudo fora. Em incidente, **restaurar o serviço vem antes de entender tudo**.
