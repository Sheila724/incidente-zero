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

**Requisitos:** Linux ou macOS com Docker. No Windows, use o WSL2. Em Linux, o [Sysbox](https://github.com/nestybox/sysbox) é recomendado para isolar os labs.

```bash
curl -fsSL https://raw.githubusercontent.com/Sheila724/incidente-zero/main/install.sh | sh

iz pull       # baixa a imagem base dos labs
iz doctor     # confere Docker, runtime e imagem
iz ui         # abre em http://127.0.0.1:8080
```

O instalador baixa o binário da última versão, confere o checksum e instala em `/usr/local/bin` (ou em `~/.local/bin`). Os cenários vêm **embutidos no binário**: não precisa clonar o repositório. Para uma versão específica, use `IZ_VERSION=v0.1.0`.

Prefere baixar na mão? Os binários de cada versão ficam em [Releases](https://github.com/Sheila724/incidente-zero/releases).

### Compilando do código

Para contribuir ou testar mudanças, compile a partir do repositório (precisa de Go 1.21+, que baixa sozinho a versão definida no `go.mod`, e Node 20.19+):

```bash
git clone https://github.com/Sheila724/incidente-zero.git
cd incidente-zero
make build        # compila a interface e o binário
make test         # roda os testes, inclusive a validação de todos os cenários
./iz build        # constrói a imagem base localmente
./iz ui
```

Rodando de dentro do repositório, o `iz` usa a pasta `scenarios/` local, então suas mudanças nos cenários aparecem na hora.

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
- **Aviso de contorno**: se você esconder o sintoma sem corrigir a causa (rollback, desligar o serviço, remover uma proteção), ou se faltar só parte da correção, a verificação avisa e pede para continuar tentando.

### Zabbix de verdade nos labs

Os cenários 014 e 015 rodam um **Zabbix 7.0 LTS completo dentro do lab** (server, frontend em português e agente 2). O botão **Zabbix**, no topo do lab, abre o frontend pela própria interface do Incidente Zero (usuário `Admin`, senha `zabbix`). O incidente começa como no dia a dia de um NOC: com um alerta no Zabbix.

No terminal do lab, `zbx-api` consulta a API do Zabbix, por exemplo: `zbx-api problem.get '{"output":["name","severity"]}'`.

Esses cenários usam a imagem `lab-zabbix` (maior que a base) e precisam de cerca de 1,5 GB de RAM livres.

### Kubernetes de verdade nos labs

Os cenários da trilha avançada rodam um **Kubernetes de nó único (k3s) dentro do lab**, com a Loja Exemplo publicada num Deployment, um Service e o nginx do servidor na frente. O `kubectl` já vem configurado no namespace `loja` (atalho `k`, com autocompletar), e uma registry interna (`registry.loja.local:5000`) responde à listagem de tags como uma registry de verdade.

Tudo funciona sem internet: as imagens do cluster já vêm dentro da imagem `lab-k3s`. Cada lab precisa de cerca de 1 GB de RAM livre (o limite é 2 GB) e leva uns 30 segundos a mais para subir, o tempo do cluster ficar pronto.

### Modo Plantão

Em **Plantão**, no topo da interface, você assume um turno (30 min, 1 h, 2 h ou 4 h), escolhe o ritmo (calmo, normal ou intenso) e as trilhas. Os incidentes caem sem aviso, em horários sorteados, um por vez:

- **Alerta de verdade:** notificação do navegador, alarme sonoro e um alerta na tela, em qualquer página, até você clicar em **Reconhecer**.
- **Escalonamento:** sem reconhecimento em 5 minutos, o incidente é escalado e perde 20 pontos.
- **Sem modo guiado:** o lab abre como num plantão real; as dicas continuam disponíveis, com custo.
- **Relatório do turno:** incidentes atendidos, MTTA (tempo até reconhecer), MTTR, escalonamentos e pontos.

Deixe a aba aberta durante o turno. Se o turno acabar com você trabalhando num incidente reconhecido, dá para terminar o que começou.

### Postmortem guiado

Depois de resolver, o botão **Escrever postmortem** abre um documento já com a **linha do tempo do seu plantão**: quando o alerta abriu, cada verificação, cada dica e a resolução. Você completa as seções de um postmortem sem culpados:

- resumo e causa raiz;
- **evidência que confirmou a causa**;
- **o que só coincidiu com a recuperação** (um restart, um deploy próximo);
- ações de prevenção, uma por linha (viram uma checklist).

O resultado sai em Markdown, para baixar ou copiar para um wiki ou ticket, e dá para comparar com o postmortem de referência do cenário. Pelo terminal, `iz postmortem <cenário>` mostra o documento.

### Pontuação

| Regra | Pontos |
|---|---|
| Base de cada incidente | 100 |
| Cada dica revelada | −10, −20 ou −40 |
| Resolvido dentro do tempo bônus do cenário | +10 |
| Solução aberta antes de resolver | 0 na tentativa |

O progresso fica salvo em `~/.config/incidente-zero/progress.json`.

### Num servidor

`iz ui` escuta só em `127.0.0.1`. Para acessar uma VPS do seu computador, use um túnel SSH:

```bash
ssh -L 8080:127.0.0.1:8080 usuario@seu-servidor
```

Para deixar a interface sempre no ar, use o serviço do systemd em [`deploy/incidente-zero.service`](deploy/incidente-zero.service). Ele roda com `--lab-idle 30m`, que remove sozinho os labs esquecidos: um lab sem a página aberta por 30 minutos é destruído.

### Publicando com HTTPS e senha

A interface dá acesso a terminais, então **nunca exponha o `iz ui` direto na internet**. Em [`deploy/traefik/`](deploy/traefik/) há um Traefik pronto, com certificado do Let's Encrypt e basic auth, que encaminha para o `iz ui` em `127.0.0.1`:

```bash
cd deploy/traefik
openssl passwd -apr1                 # gera o hash da senha
nano dynamic.yml                     # domínio e usuário:hash
nano docker-compose.yml              # e-mail do Let's Encrypt
docker compose up -d
```

O domínio precisa apontar para o servidor. Sem domínio próprio, use o [sslip.io](https://sslip.io): `iz.203-0-113-10.sslip.io` resolve para `203.0.113.10`.

## Pelo terminal

Tudo também funciona sem a interface:

```bash
iz list         # lista os cenários
iz start 001    # sobe o lab e abre o terminal do incidente
iz check        # depois de sair do lab (exit): valida a correção
iz hint 1       # dica, se travar
iz solution     # solução comentada e causa raiz
iz stop         # destrói o lab
```

| Comando | O que faz |
|---|---|
| `iz ui` | Abre a interface web |
| `iz pull` | Baixa a imagem base pronta |
| `iz doctor` | Confere Docker, runtime, imagem e cenários |
| `iz list` | Lista os cenários disponíveis |
| `iz build` | Constrói a imagem base localmente |
| `iz start <cenário>` | Sobe o lab, injeta o incidente e abre o terminal |
| `iz shell` | Reabre o terminal do lab |
| `iz check` | Verifica se o incidente foi resolvido |
| `iz hint [n]` | Mostra a dica n |
| `iz concepts` | Conceitos para estudar antes |
| `iz solution` | Solução comentada, causa raiz e referências |
| `iz postmortem` | Mostra o seu postmortem em Markdown |
| `iz stop` | Destrói o lab |
| `iz selftest [cenários]` | Testa os cenários de ponta a ponta (para quem cria cenários) |

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

### Trilha intermediária: rede, web, banco e Zabbix

| ID | Incidente | O que você pratica | Status |
|---|---|---|---|
| 009 | Checkout falhando após migração da API de pagamentos | `getent` × `dig`, `/etc/hosts`, `nsswitch.conf`, cache de DNS na aplicação | ✅ |
| 010 | Certificado TLS expirado em produção | `openssl s_client`, CA interna, `reload` do nginx, timers do systemd | ✅ |
| 011 | 502 Bad Gateway após deploy | proxy reverso, `upstream`, `error.log`, `ss -ltnp` | ✅ |
| 012 | Banco de dados recusando conexões | `SHOW PROCESSLIST`, `KILL`, vazamento de conexões, `MAX_USER_CONNECTIONS` | ✅ |
| 013 | Banco travado por disco cheio | binlog, `PURGE BINARY LOGS`, expiração, disco cheio no MariaDB | ✅ |
| 014 | Monitoramento cego: host sem dados no Zabbix | `zabbix_get`, `Server=`/`ServerActive=`, log do agente, falha silenciosa | ✅ |
| 015 | Memória crescendo sem parar após deploy | gráficos do Zabbix, RSS, vazamento, `MemoryMax`, registrar o incidente no Zabbix | ✅ |

### Trilha avançada: Kubernetes e incidentes em cascata

| ID | Incidente | O que você pratica | Status |
|---|---|---|---|
| 016 | Loja fora do ar: pods em CrashLoopBackOff | `kubectl logs --previous`, `describe`, ConfigMap, `rollout restart` | ✅ |
| 017 | Deploy travado em ImagePullBackOff | rolling update, ReplicaSets, eventos do pod, tags na registry | ✅ |
| 018 | Pods rodando, loja fora do ar: Service sem endpoints | readiness probe, Service e EndpointSlices, `kubectl exec` | ✅ |
| 019 | Checkout novo preso em Pending | scheduler, requests e limits, unidades de CPU, `describe node` | ✅ |
| 020 | Madrugada: a loja cai, volta e cai de novo | falha em cascata, priorização, CronJobs e Jobs, labels e seletores | ✅ |

## Roadmap

- [x] Imagem base com systemd
- [x] CLI `iz`
- [x] Trilha iniciante completa (001–008)
- [x] Interface web com terminal no navegador e monitoramento ao vivo
- [x] Pontuação, MTTR pessoal e histórico de plantões
- [x] Trilha intermediária: rede, web, banco e **Zabbix**
- [x] Trilha avançada: **Kubernetes** (k3s) e incidentes em cascata
- [x] Postmortem guiado, com a linha do tempo do seu plantão
- [x] Binários prontos para download, instalador e CI
- [x] Demo online com HTTPS e senha, e limpeza automática de labs
- [ ] Variações nos cenários (nomes, portas e serviços diferentes a cada execução)
- [x] **Modo Plantão:** turno com incidentes sorteados, alerta com som e notificação, reconhecimento, escalonamento e relatório do turno
- [ ] Modo Plantão: escala semanal e alertas no celular

## Arquitetura

```
Navegador ── HTTP / WebSocket ──▶ iz ui (Go, interface embutida)
                                   ├─ API: cenários, progresso, verificação, dicas, métricas
                                   ├─ Terminal: WebSocket ⇄ PTY ⇄ docker exec
                                   └─ Docker ──▶ lab (Ubuntu + systemd, isolado com Sysbox)
```

- **Labs:** containers Ubuntu com systemd, cron, SSH e MariaDB (desligado por padrão), mais as variantes com Zabbix (`lab-zabbix`) e Kubernetes (`lab-k3s`). No Linux, usa o runtime [Sysbox](https://github.com/nestybox/sysbox) para isolamento forte sem `--privileged`. Sem Sysbox, roda no modo local. Cada lab tem limite de CPU, memória e processos, e os incidentes de disco usam partições `tmpfs` próprias: nunca enchem o seu disco.
- **Backend:** Go, um binário só, com a interface embutida (`go:embed`). O terminal usa WebSocket e PTY, e só aceita conexões da mesma origem.
- **Interface:** Svelte 5 + Vite, xterm.js para o terminal e fontes locais, sem chamadas externas.
- **Cenários:** cada um é uma pasta com `scenario.yaml`, conteúdo em Markdown e scripts `setup`, `inject` e `check`. Adicionar um cenário não exige mexer em código.
- **Interfaces dentro do lab:** um cenário pode declarar `apps` (como o frontend do Zabbix). A interface do iz publica cada uma em `/apps/<cenário>/<nome>/`, atrás da mesma autenticação.

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

## Publicando uma versão

Uma tag `v*` dispara o workflow de release, que roda os testes, compila os binários para Linux e macOS (amd64 e arm64), publica a release com os checksums e envia as imagens dos labs (`lab-base`, `lab-zabbix` e `lab-k3s`) para o GHCR. Antes de criar a tag, atualize o `main` local (`git pull`) e confira o commit com `git log --oneline -1`:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Contribuindo

Quer criar um cenário? O passo a passo completo está no [CONTRIBUTING.md](CONTRIBUTING.md). Em resumo:

- **O `inject.sh` termina com uma autoverificação.** Se a falha não aparecer, o lab não começa.
- **O `check.sh` confere a causa raiz**, não só o sintoma, e não aceita "soluções" que criam outro problema (como `chmod 777` ou desligar verificações de segurança).
- **Cada cenário tem uma solução de referência** (`scripts/solve.sh`), e `iz selftest <cenário>` testa tudo de ponta a ponta: a falha aparece, o check reprova e, depois da solução, aprova.

O CI valida todos os cenários (`go test ./...`, `shellcheck`) e, em cada PR, roda o `selftest` dos cenários alterados num lab de verdade. Toda segunda-feira, todos os cenários são testados de novo.

Toda contribuição de incidente real que você já viveu, anonimizado, é bem-vinda.

## Projetos relacionados

- [NOC Horizon](https://github.com/Sheila724/noc-horizon): painel de monitoramento em tempo real em cima do Zabbix, com SLO, MTTR e MTTA. O NOC Horizon mostra que algo quebrou; o Incidente Zero treina o que fazer depois do alerta.

## Licença

[MIT](LICENSE) © Sheila Alves Araujo
