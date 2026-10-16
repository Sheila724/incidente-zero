# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja o que está sendo servido

```bash
curl -s http://localhost | head -20
```

```html
<title>Directory listing for /</title>
<li><a href="bin/">bin@</a></li>
<li><a href="etc/">etc/</a></li>
```

Isso é a raiz do servidor sendo exposta. Qualquer um consegue baixar, por exemplo, `http://loja-web-01/etc/passwd`. Prioridade máxima: tirar isso do ar.

## 2. O nginx

```bash
systemctl status nginx --no-pager
journalctl -u nginx -n 20 --no-pager
```

```
bind() to 0.0.0.0:80 failed (98: Address already in use)
```

O nginx não subiu porque alguém chegou primeiro na porta 80.

## 3. Quem ocupa a porta

```bash
sudo ss -ltnp 'sport = :80'
```

```
LISTEN 0 5 0.0.0.0:80 0.0.0.0:* users:(("python3",pid=212,fd=3))
```

## 4. De onde ele veio

```bash
systemctl status 212
systemctl cat debug-http
```

```ini
# Servidor de teste — subir só durante a investigação do chamado #4512
[Service]
ExecStart=/usr/bin/python3 -m http.server 80 --directory /
Restart=always

[Install]
WantedBy=multi-user.target
```

Alguém criou um servidor de teste, **habilitou** e esqueceu. Enquanto o nginx já estava rodando, ele falhava em silêncio. No reboot, chegou primeiro na porta.

## 5. Pare e desabilite

```bash
sudo systemctl disable --now debug-http
```

Só `kill` não basta: há `Restart=always`. Só `stop` também não: no próximo reboot ele volta.

## 6. Suba o nginx

```bash
sudo systemctl start nginx
curl -s http://localhost | grep h1
```

## 7. Depois do incidente

Como a raiz do servidor ficou exposta por horas, trate também como **incidente de segurança**: revise os logs de acesso, verifique se há chaves ou senhas em arquivos que estavam acessíveis e faça a rotação delas.
