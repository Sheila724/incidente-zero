# 🚨 Incidente Zero

> Aprenda a resolver incidentes antes do seu primeiro plantão de verdade.

O **Incidente Zero** é um simulador de incidentes open source para quem está começando em **DevOps, SRE e Infraestrutura**. Em vez de seguir um tutorial, você recebe um alerta, entra num servidor quebrado e precisa diagnosticar, corrigir a causa raiz e entender o que aconteceu.

Tudo em português, do iniciante ao avançado, rodando na sua máquina.

## Como funciona

1. Você assume um plantão e recebe o alerta: *"Loja Exemplo retornando HTTP 503"*.
2. Um lab isolado sobe com a falha já injetada: um Linux real com systemd, nginx e a aplicação.
3. Você investiga pelo terminal com as ferramentas de verdade: `df`, `journalctl`, `systemctl`, `ss`, `curl`.
4. A verificação confere se você corrigiu a **causa raiz**, não só o sintoma.
5. Cada cenário traz conceitos, dicas progressivas, solução comentada e uma análise de causa raiz no formato de postmortem.

Todo cenário tem uma armadilha: matar o processo que volta sozinho, apagar o log e ver o disco encher de novo, corrigir o primeiro erro e descobrir um segundo. É assim que acontece num plantão de verdade.

## Início rápido

**Requisitos:** Linux com Docker (Windows via WSL2 e Mac também funcionam no modo local), Go 1.22+ e Node 20.19+ para compilar.

```bash
git clone https://github.com/Sheila724/incidente-zero.git
cd incidente-zero
make build        # compila a interface web e o binário iz
./iz build        # constrói a imagem base dos labs
./iz ui           # abre em http://127.0.0.1:8080
```

Depois de atualizar o projeto (`git pull`), rode `make build` e `./iz build` de novo: a imagem base também evolui junto com os cenários.

## A interface

### Central de plantão
- Seus números: incidentes resolvidos, **MTTR pessoal**, dicas usadas e pontuação.
- O próximo incidente recomendado, com o alerta e o que você vai praticar.
- Progresso por trilha e catálogo de cenários, com o status de cada um.
- Histórico de plantões, com tempo, dicas e pontos.

### Lab do incidente
- Barra do incidente com severidade e cronômetro do plantão.
- **Terminal real no navegador**, conectado ao servidor do lab.
- **Modo guiado**: os passos são verificados automaticamente enquanto você trabalha.
- Dicas progressivas, conceitos para estudar e, no fim, a solução e a causa raiz.
- **Monitoramento ao vivo**: HTTP, porta 80, CPU, memória, disco e inodes, que mudam enquanto você corrige.
- Botão **Verificar** (`Ctrl+Enter`) e, depois de resolver, o atalho para o **próximo incidente**.

### Pontuação

| Regra | Pontos |
|---|---|
| Base de cada incidente | 100 |
| Cada dica revelada | −10, −20 ou −40 |
| Resolvido dentro do tempo bônus do cenário | +10 |
| Solução aberta antes de resolver | 0 na tentativa |

O progresso fica salvo em `~/.config/incidente-zero/progress.json`.

### Num servidor

`./iz ui` escuta só em `127.0.0.1`. Para acessar uma VPS do seu computador, use um túnel SSH:

```bash
ssh -L 8080:127.0.0.1:8080 usuario@seu-servidor
```

Para expor de verdade, use `./iz ui --addr 0.0.0.0:8080` **somente atrás de um proxy com autenticação** (Traefik com basic auth, por exemplo): a interface dá acesso a terminais.

## Pelo terminal

Tudo também funciona sem a interface:

```bash
./iz list         # lista os cenários
./iz start 001    # sobe o lab e abre o terminal do incidente
./iz check        # depois de sair do lab (exit): valida a correção
./iz hint 1       # dica, se travar
./iz solution     # solução comentada e causa raiz
./iz stop         # destrói o lab
```

| Comando | O que faz |
|---|---|
| `iz ui` | Abre a interface web |
| `iz list` | Lista os cenários disponíveis |
| `iz build` | Constrói a imagem base dos labs |
| `iz start <cenário>` | Sobe o lab, injeta o incidente e abre o terminal |
| `iz shell` | Reabre o terminal do lab |
| `iz check` | Verifica se o incidente foi resolvido |
| `iz hint [n]` | Mostra a dica n |
| `iz concepts` | Conceitos para estudar antes |
| `iz solution` | Solução comentada, causa raiz e referências |
| `iz stop` | Destrói o lab |

## Cenários

### Trilha iniciante: Linux e troubleshooting

| ID | Incidente | O que você pratica | Status |
|---|---|---|---|
| 001 | Disco cheio em `/var/log` | `df`, `du`, níveis de log, `truncate`, logrotate | ✅ |
| 002 | Nginx não sobe após deploy | `systemctl`, `journalctl`, `nginx -t`, `default_server` | ✅ |
| 003 | Permissão negada após hardening | `chmod`, `chown`, `namei`, timers do systemd | ✅ |
| 004 | Processo consumindo 100% de CPU | `top`, `ps`, `Restart=always`, `CPUQuota` | ✅ |
| 005 | Porta 80 ocupada por outro processo | `ss`, `lsof`, `disable --now`, exposição de arquivos | ✅ |
| 006 | Backup agendado que não executa | cron, regras do `/etc/cron.d`, `%` na linha do cron | ✅ |
| 007 | SSH recusando a chave do deploy | `ssh -v`, log do sshd, `StrictModes`, permissões de `~/.ssh` | ✅ |
| 008 | "Disco cheio" com espaço sobrando | inodes, `df -i`, `find -delete` seletivo | ✅ |

### Próximas trilhas

| Trilha | Foco | Status |
|---|---|---|
| Intermediária | Rede, web, banco e alertas reais do **Zabbix** | 🚧 |
| Avançada | **Kubernetes** (k3s) e incidentes em cascata | 🚧 |

## Roadmap

- [x] Imagem base com systemd
- [x] CLI `iz`
- [x] Trilha iniciante completa (001–008)
- [x] Interface web com terminal no navegador e monitoramento ao vivo
- [x] Pontuação, MTTR pessoal e histórico de plantões
- [ ] Trilha intermediária: rede, web, banco e **Zabbix**
- [ ] Trilha avançada: **Kubernetes** (k3s) e incidentes em cascata
- [ ] Postmortem guiado, gerado a partir do que você fez no lab
- [ ] Binários prontos para download (sem precisar compilar)

## Arquitetura

```
Navegador ── HTTP / WebSocket ──▶ iz ui (Go, interface embutida)
                                   ├─ API: cenários, progresso, verificação, dicas, métricas
                                   ├─ Terminal: WebSocket ⇄ PTY ⇄ docker exec
                                   └─ Docker ──▶ lab (Ubuntu + systemd, isolado com Sysbox)
```

- **Labs:** containers Ubuntu com systemd, cron e SSH. No Linux, usa o runtime [Sysbox](https://github.com/nestybox/sysbox) para isolamento forte sem `--privileged`. Sem Sysbox, roda no modo local. Cada lab tem limite de CPU, memória e processos, e os incidentes de disco usam partições `tmpfs` próprias: nunca enchem o seu disco.
- **Backend:** Go, um binário só, com a interface embutida (`go:embed`). O terminal usa WebSocket e PTY, e só aceita conexões da mesma origem.
- **Interface:** Svelte 5 + Vite, xterm.js para o terminal e fontes locais, sem chamadas externas.
- **Cenários:** cada um é uma pasta com `scenario.yaml`, conteúdo em Markdown e scripts `setup`, `inject` e `check`. Adicionar um cenário não exige mexer em código.

```
scenarios/iniciante/001-disco-cheio/
├── scenario.yaml     # metadados, recursos, alerta e passos do modo guiado
├── briefing.md       # o alerta que abre o incidente
├── conceitos.md      # o que estudar antes
├── hints.yaml        # dicas progressivas
├── solucao.md        # solução comentada
├── causa-raiz.md     # linha do tempo, 5 porquês e prevenção
├── referencias.md    # leitura extra
└── scripts/
    ├── setup.sh      # ambiente saudável
    ├── inject.sh     # injeta a falha e confere que ela apareceu
    └── check.sh      # exit 0 = resolvido; uma linha ✅/❌ por verificação
```

## Contribuindo

Quer criar um cenário? Use o `001-disco-cheio` como modelo. Algumas regras que mantêm a qualidade:

- **O `inject.sh` termina com uma autoverificação.** Se a falha não aparecer, o lab não começa.
- **O `check.sh` confere a causa raiz**, não só o sintoma, e não aceita "soluções" que criam outro problema (como `chmod 777` ou desligar verificações de segurança).
- **Partições de teste com `tmpfs`** nos incidentes de disco (`size` e, se precisar, `inodes` em `resources.mounts`).
- **Passos do modo guiado com `verify`** sempre que der para verificar automaticamente.

Toda contribuição de incidente real que você já viveu, anonimizado, é bem-vinda.

## Projetos relacionados

- [NOC Horizon](https://github.com/Sheila724/noc-horizon): painel de monitoramento em tempo real em cima do Zabbix, com SLO, MTTR e MTTA. O NOC Horizon mostra que algo quebrou; o Incidente Zero treina o que fazer depois do alerta.

## Licença

[MIT](LICENSE) © Sheila Alves Araujo
