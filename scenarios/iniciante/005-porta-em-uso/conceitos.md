# Conceitos para este cenário

## Portas e "escutar"

Um serviço de rede **escuta** (*listen*) numa porta: 80 para HTTP, 443 para HTTPS, 22 para SSH. Só **um** processo pode escutar numa mesma porta e endereço por vez. Se outro tentar, recebe:

```
bind() to 0.0.0.0:80 failed (98: Address already in use)
```

Portas abaixo de 1024 exigem privilégios de root para abrir.

## Quem está na porta?

```bash
sudo ss -ltnp                 # todas as portas TCP escutando, com processo
sudo ss -ltnp 'sport = :80'   # só a porta 80
sudo lsof -i :80              # alternativa
sudo fuser 80/tcp             # só o PID
```

Em `ss -ltnp`: `-l` escutando, `-t` TCP, `-n` números em vez de nomes, `-p` processo. Sem `sudo`, a coluna do processo aparece vazia para processos de outros usuários.

## Do PID ao serviço

```bash
systemctl status 1234     # qual unit iniciou esse PID
systemctl cat unit        # o arquivo da unit
```

## Parar x desabilitar

| Comando | Agora | No próximo boot |
|---|---|---|
| `systemctl stop x` | Para | Volta, se estiver habilitado |
| `systemctl disable x` | Continua rodando | Não sobe |
| `systemctl disable --now x` | Para | Não sobe |
| `systemctl mask x` | Não pode ser iniciado nem manualmente | Não sobe |

Um incidente que aparece **depois de um reboot** quase sempre envolve algo habilitado (`enabled`) que ninguém lembrava.

## Ordem de inicialização

No boot, o systemd sobe vários serviços em paralelo. Se dois querem a mesma porta, vence quem chegar primeiro, e o outro falha. Por isso o problema pode ficar escondido até o próximo reboot.

## `python3 -m http.server`

Um servidor web de uma linha, ótimo para testes rápidos. Ele **lista e entrega qualquer arquivo** da pasta onde roda (ou de `--directory`). Rodando como root e servindo `/`, entrega o servidor inteiro para quem acessar, inclusive arquivos de configuração e chaves.
