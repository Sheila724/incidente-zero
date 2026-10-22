# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Ontem, 18:30 | O time de dados aplica o manifesto novo do `loja-db`, copiado de homologação, com `ambiente=homologacao` |
| Ontem, 18:32 | A CronJob `economia-homologacao` desliga o `loja-db` pela primeira vez |
| Ontem à noite | Alguém sobe o banco manualmente várias vezes; ele volta a cair a cada 2 minutos |
| 02:10 | O banco fica desligado, a readiness da loja falha, o Service fica sem endpoints e o alerta dispara |

## Os 5 porquês

1. **Por que a loja respondia 502?** O Service `loja-web` não tinha pods prontos.
2. **Por que os pods não estavam prontos?** A readiness probe não alcançava o banco `loja-db`.
3. **Por que o banco não respondia?** O Deployment estava com zero réplicas.
4. **Por que zero réplicas?** A CronJob de economia desliga tudo com `ambiente=homologacao`, e o banco tinha esse rótulo.
5. **Por que o banco tinha esse rótulo?** O manifesto de produção foi copiado de homologação e só a imagem e as réplicas foram revisadas.

## Fatores que contribuíram

- Uma automação que age por rótulo, num cluster sem validação de rótulos.
- Mitigações repetidas (subir o banco) sem investigar por que ele caía.
- A cascata escondeu a origem: o alerta era da loja, o problema era do banco, e a causa era de uma terceira peça.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Manifestos por ambiente gerados de uma base comum (Kustomize/Helm), não copiados à mão | Prevenção |
| Produção e homologação em namespaces (ou clusters) separados | Prevenção |
| Política que impede `ambiente=homologacao` no namespace de produção | Prevenção |
| A automação registra (e avisa) o que desliga | Detecção |
| Alerta para Deployment com zero réplicas desejadas em produção | Detecção |

## Evidência × coincidência

| O que foi observado | Peso |
|---|---|
| `Readiness probe failed: banco loja-db:3306 inacessível` | Evidência do caminho da cascata |
| `loja-db` com 0 réplicas desejadas e o evento `Scaled down ... to 0` | Evidência de que alguém desligou o banco |
| Log do job: `desligando loja-db` e o rótulo `ambiente=homologacao` no banco | Evidência da causa |
| A loja voltar depois de subir o banco | Mitigação: o problema volta no próximo ciclo |
| O horário de madrugada | Coincidência: a CronJob roda o dia todo |

## Para o seu postmortem

Quando o incidente volta sozinho, a pergunta muda de "o que quebrou?" para "quem está agindo?". Automações fazem exatamente o que foram programadas para fazer, inclusive com os objetos errados.
