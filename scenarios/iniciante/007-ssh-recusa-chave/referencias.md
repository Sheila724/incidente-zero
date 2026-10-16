# Para estudar mais

## SSH

- [sshd_config(5)](https://man.openbsd.org/sshd_config): veja `StrictModes`, `AuthorizedKeysFile` e `PasswordAuthentication`
- [sshd(8)](https://man.openbsd.org/sshd): seção sobre o arquivo `authorized_keys`
- [ssh(1)](https://man.openbsd.org/ssh): a opção `-v` para depurar

## Permissões

- [namei](https://man7.org/linux/man-pages/man1/namei.1.html): permissões de cada parte do caminho
- [chmod](https://man7.org/linux/man-pages/man1/chmod.1.html) e [chown](https://man7.org/linux/man-pages/man1/chown.1.html)

## Para praticar

- Gere um par de chaves com `ssh-keygen -t ed25519` e configure login por chave numa VM.
- Rode `chmod 777 ~/.ssh` de propósito nessa VM e leia o log do servidor.
- Cenário relacionado: **003 — Permissão negada após hardening**.
