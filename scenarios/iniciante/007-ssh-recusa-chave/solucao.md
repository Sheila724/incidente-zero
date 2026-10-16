# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Reproduza

```bash
ssh -i ~/ci/deploy_key deploy@localhost
```

```
deploy@localhost: Permission denied (publickey).
```

## 2. O cliente oferece a chave?

```bash
ssh -v -i ~/ci/deploy_key deploy@localhost 2>&1 | grep -iE 'offering|authentications'
```

A chave é oferecida, mas o servidor não aceita. O problema está no servidor.

## 3. O motivo real

```bash
journalctl -u ssh -n 30 --no-pager
```

```
Authentication refused: bad ownership or modes for file /home/deploy/.ssh/authorized_keys
```

## 4. Veja o caminho

```bash
namei -l /home/deploy/.ssh/authorized_keys
ls -la /home/deploy /home/deploy/.ssh
```

```
drwxrwxrwx deploy deploy deploy
drwxrwxrwx deploy deploy .ssh
-rwxrwxrwx root   root   authorized_keys
```

Dois problemas: o `chmod -R 777` deixou tudo gravável por todos, e o `authorized_keys` ficou com dono `root`, porque a chave nova foi copiada com `sudo cp` na rotação.

## 5. Corrija

```bash
sudo chmod 755 /home/deploy
sudo chmod 700 /home/deploy/.ssh
sudo chmod 600 /home/deploy/.ssh/authorized_keys
sudo chown -R deploy:deploy /home/deploy/.ssh
```

Se você corrigir só as permissões e esquecer o dono, o log muda para:

```
Could not open user 'deploy' authorized keys '/home/deploy/.ssh/authorized_keys': Permission denied
```

O sshd lê esse arquivo **como o usuário** `deploy`, e um arquivo `600` do root ele não consegue ler.

## 6. Valide

```bash
ssh -i ~/ci/deploy_key deploy@localhost 'cat app/release.txt'
```

## 7. Bônus: desfaça o resto do 777

```bash
sudo find /home/deploy -perm -o+w
sudo find /home/deploy -type d -exec chmod 755 {} +
sudo find /home/deploy -type f -exec chmod 644 {} +
sudo chmod 700 /home/deploy/.ssh
sudo chmod 600 /home/deploy/.ssh/authorized_keys
sudo chmod 755 /home/deploy/app/deploy.sh
```

Atenção ao último comando: scripts que precisam executar continuam com `x`.
