# Conceitos para este cenário

## Load average

```bash
uptime
# 15:44:02 up 3 days,  load average: 1.92, 1.85, 1.40
```

São as médias de processos querendo CPU no último 1, 5 e 15 minutos. Compare com a quantidade de CPUs disponíveis (`nproc`): load bem acima disso significa fila para usar o processador.

## `top`: o painel de controle

| Tecla | O que faz |
|---|---|
| `P` | Ordena por CPU (padrão) |
| `M` | Ordena por memória |
| `c` | Mostra o comando completo |
| `1` | Mostra cada CPU separadamente |
| `q` | Sai |

Colunas importantes: `PID`, `USER`, `%CPU`, `%MEM`, `COMMAND`.

## `ps`: uma foto dos processos

```bash
ps aux --sort=-%cpu | head      # os que mais usam CPU
ps -o pid,ppid,user,%cpu,cmd -p 1234
```

O `%CPU` do `ps` é a **média desde que o processo começou**. Para o uso de agora, prefira o `top`.

## Quem iniciou esse processo?

Quase tudo num servidor moderno é iniciado pelo systemd. Dado um PID:

```bash
systemctl status 1234      # mostra a unit dona do processo
systemctl cat nome-do-servico
```

## `Restart=always`

Na unit de um serviço, `Restart=always` manda o systemd **subir o processo de novo** sempre que ele morrer. É ótimo para resiliência, mas significa que `kill` não "desliga" nada: o processo volta com outro PID em segundos.

Para controlar um serviço, use o systemd:

```bash
sudo systemctl stop servico      # para agora
sudo systemctl restart servico   # reinicia (relê a configuração)
sudo systemctl disable servico   # não sobe no boot
```

## Sinais: `kill`

| Sinal | Número | Efeito |
|---|---|---|
| `SIGTERM` | 15 | Pede para encerrar (padrão do `kill`) |
| `SIGKILL` | 9 | Encerra à força, sem chance de limpar nada |

Use o `-9` só quando o `TERM` não funcionar.

## Limitando CPU: `nice` e `CPUQuota`

- `nice` / `renice` mudam a **prioridade**: o processo continua usando CPU livre, mas cede para os outros.
- `CPUQuota=20%` numa unit do systemd é um **limite rígido**: o serviço nunca passa de 20% de um núcleo.

```bash
sudo systemctl edit servico   # cria um override sem mexer no arquivo original
```
