# Para estudar mais

## Rede

- [ss](https://man7.org/linux/man-pages/man8/ss.8.html): portas e conexões, com filtros
- [lsof](https://man7.org/linux/man-pages/man8/lsof.8.html): arquivos e sockets abertos por processo
- [fuser](https://man7.org/linux/man-pages/man1/fuser.1.html): quem está usando um arquivo ou porta

## systemd

- [systemctl](https://www.freedesktop.org/software/systemd/man/latest/systemctl.html): `enable`, `disable --now`, `mask`, `list-unit-files`

## Segurança

- [OWASP: Directory listing](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/02-Configuration_and_Deployment_Management_Testing/04-Review_Old_Backup_and_Unreferenced_Files_for_Sensitive_Information): por que expor arquivos é grave

## Para praticar

- Rode `python3 -m http.server 8000` numa pasta qualquer e veja o resultado com `ss -ltnp`.
- Tente subir um segundo servidor na mesma porta e leia o erro.
- Liste os serviços habilitados da sua máquina: `systemctl list-unit-files --state=enabled`.
