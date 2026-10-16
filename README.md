# 🚨 Incidente Zero

> Aprenda a resolver incidentes antes do seu primeiro plantão de verdade.

O **Incidente Zero** é um simulador de incidentes open source para quem está começando em **DevOps, SRE e Infraestrutura**. Em vez de seguir um tutorial, você recebe um alerta, entra num servidor quebrado e precisa diagnosticar, corrigir a causa raiz e entender o que aconteceu.

Tudo em português, do iniciante ao avançado, rodando na sua máquina.

## Como funciona

1. Você inicia um cenário e recebe o alerta: *"Loja Exemplo retornando HTTP 503"*.
2. Um lab isolado sobe com a falha já injetada: um Linux real com systemd, nginx e a aplicação.
3. Você investiga pelo terminal com as ferramentas de verdade: `df`, `journalctl`, `systemctl`, `curl`.
4. O `iz check` valida se você corrigiu a **causa raiz**, não só o sintoma.
5. Cada cenário traz conceitos, dicas progressivas, solução comentada e uma análise de causa raiz no formato de postmortem.

## Início rápido

**Requisitos:** Linux com Docker (Windows via WSL2 e Mac também funcionam no modo local) e Go 1.22+ para compilar.

```bash
git clone https://github.com/Sheila724/incidente-zero.git
cd incidente-zero
go build -o iz ./cmd/iz

./iz build        # constrói a imagem base dos labs
./iz list         # lista os cenários
./iz start 001    # sobe o lab e abre o terminal do incidente
```

Dentro do lab, investigue e corrija. Depois:

```bash
./iz check        # valida a correção
./iz hint 1       # dica, se travar
./iz solution     # solução comentada e causa raiz
./iz stop         # destrói o lab
```

## Comandos

| Comando | O que faz |
|---|---|
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

| ID | Nível | Trilha | Incidente | Status |
|---|---|---|---|---|
| 001 | Iniciante | Linux | Disco cheio em `/var/log` | ✅ |
| 002 | Iniciante | Linux | Nginx não sobe | 🚧 |
| 003 | Iniciante | Linux | Permissão negada no site | 🚧 |
| 004 | Iniciante | Linux | Processo consumindo 100% de CPU | 🚧 |
| 005 | Iniciante | Linux | Porta já em uso | 🚧 |
| 006 | Iniciante | Linux | Cron que não executa | 🚧 |
| 007 | Iniciante | Linux | SSH recusando a chave | 🚧 |
| 008 | Iniciante | Linux | "Disco cheio" com espaço sobrando | 🚧 |

## Roadmap

- [x] Imagem base com systemd
- [x] Cenário 001 completo
- [x] CLI `iz`
- [ ] Trilha iniciante completa (001–008)
- [ ] Interface web com terminal no navegador
- [ ] Trilha intermediária: rede, web, banco e **Zabbix**
- [ ] Trilha avançada: **Kubernetes** (k3s) e incidentes em cascata
- [ ] Modo plantão com pontuação e postmortem guiado

## Arquitetura

- **Labs:** containers Ubuntu com systemd. No Linux, usa o runtime [Sysbox](https://github.com/nestybox/sysbox) para isolamento forte sem `--privileged`. Sem Sysbox, roda no modo local.
- **Cenários:** cada um é uma pasta com `scenario.yaml`, conteúdo em Markdown e scripts `setup`, `inject` e `check`. Adicionar um cenário não exige mexer em código.
- **CLI:** escrito em Go, um binário só.

```
scenarios/iniciante/001-disco-cheio/
├── scenario.yaml     # metadados, recursos e passos do modo guiado
├── briefing.md       # o alerta que abre o incidente
├── conceitos.md      # o que estudar antes
├── hints.yaml        # dicas progressivas
├── solucao.md        # solução comentada
├── causa-raiz.md     # linha do tempo, 5 porquês e prevenção
├── referencias.md    # leitura extra
└── scripts/          # setup.sh, inject.sh, check.sh
```

## Contribuindo

Quer criar um cenário? Use o `001-disco-cheio` como modelo. Toda contribuição de incidente real que você já viveu, anonimizado, é bem-vinda.

## Licença

[MIT](LICENSE) © Sheila Alves Araujo
