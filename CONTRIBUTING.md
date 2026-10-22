# Contribuindo com o Incidente Zero

Obrigada pelo interesse! A forma mais valiosa de contribuir é **criar um cenário**: um incidente que você já viveu (anonimizado) vira um exercício para quem está começando.

## 1. Prepare o ambiente

Você precisa de Linux (ou WSL2) com Docker, Go 1.21+ (ele baixa sozinho a versão definida no `go.mod`) e Node 20.19+.

```bash
git clone https://github.com/Sheila724/incidente-zero.git
cd incidente-zero
make build        # compila a interface e o binário ./iz
./iz pull         # baixa as imagens dos labs (ou ./iz build para construir)
./iz doctor
```

Rodando de dentro do repositório, o `./iz` usa a pasta `scenarios/` local: o que você mudar aparece na hora.

## 2. Crie a pasta do cenário

Copie um cenário parecido como ponto de partida (o `001-disco-cheio` é o mais simples):

```bash
cp -r scenarios/iniciante/001-disco-cheio scenarios/iniciante/016-meu-cenario
```

| Arquivo | Para quê |
|---|---|
| `scenario.yaml` | Metadados, recursos do lab, alerta, passos do modo guiado e pontuação |
| `briefing.md` | O alerta que abre o incidente: o que o plantão sabe no começo |
| `conceitos.md` | O que estudar antes (sem entregar a solução) |
| `hints.yaml` | Três dicas progressivas, da mais vaga à mais direta |
| `solucao.md` | Solução comentada, passo a passo, com a armadilha explicada |
| `causa-raiz.md` | Postmortem: linha do tempo, 5 porquês, fatores e prevenção |
| `referencias.md` | Leituras para aprofundar |
| `scripts/setup.sh` | Monta o ambiente **saudável** |
| `scripts/inject.sh` | Injeta a falha e confere que ela apareceu |
| `scripts/check.sh` | Valida a correção: uma linha ✅/❌ por verificação, exit 0 = resolvido. Use ⚠️ (com exit 1) quando reconhecer um contorno, como rollback, desligar o serviço ou remover uma proteção: a interface avisa a pessoa que aquilo esconde o sintoma sem corrigir a causa |
| `scripts/solve.sh` | A solução de referência, usada pelos testes automáticos |

O ID segue o padrão `NNN-nome-em-minusculas`, igual ao nome da pasta, com um número que ainda não existe.

## 3. Regras que mantêm a qualidade

- **Uma armadilha de propósito.** A primeira correção óbvia (reiniciar, apagar o arquivo, liberar a permissão para todos) não pode bastar.
- **O `check.sh` confere a causa raiz**, não só o sintoma, e recusa "soluções" que criam outro problema (`chmod 777`, desligar verificações de segurança, desativar o serviço).
- **O `check.sh` não altera o lab.** Se precisar simular algo (por exemplo, rodar um job), desfaça o que a verificação criou.
- **O `inject.sh` termina com uma autoverificação** (um comentário `# Autoverificação` seguido do teste). Se a falha não aparecer, o lab não começa.
- **Incidentes de disco usam `tmpfs`** em `resources.mounts` (com `size` e, se precisar, `inodes`): nunca enchem o disco de quem roda.
- **Scripts com `set -euo pipefail` e `umask 022`** no `setup.sh` e no `inject.sh`.
- **Passos do modo guiado com `verify`** sempre que der para verificar automaticamente.
- **Português claro**, sem jargão desnecessário, e o aluno sempre como `aluno` com `sudo`.

## 4. Teste de ponta a ponta

```bash
./iz selftest 016
```

O `selftest` sobe o lab, roda o `setup` e o `inject`, confere que o check **falha**, aplica o `solve.sh` e confere que o check **passa**. Use `--keep` para deixar o lab de pé e investigar uma falha (`./iz shell 016`).

Teste também como aluno: `./iz start 016`, resolva sem olhar a solução e caia na armadilha de propósito, para ver se o check reprova.

## 5. Antes de abrir o PR

- [ ] `go test ./...` passa (valida a estrutura de todos os cenários)
- [ ] `shellcheck -S warning scenarios/*/*/scripts/*.sh` sem avisos
- [ ] `./iz selftest <seu cenário>` passa
- [ ] A armadilha foi testada e o check reprova
- [ ] README atualizado na tabela de cenários

No PR, o GitHub Actions roda o `selftest` dos cenários que você alterou.

## Ideias e dúvidas

Use as [Discussions](https://github.com/Sheila724/incidente-zero/discussions) para propor incidentes e as issues com o label `good first issue` para começar.
