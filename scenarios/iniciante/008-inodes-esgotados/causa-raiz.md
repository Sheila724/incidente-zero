# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| Sexta, 22:00 | Migração do servidor; o timer `limpa-sessoes` é desativado "por segurança" |
| Sexta → segunda | Cada visita cria um arquivo de sessão; nada é apagado |
| Segunda, 14:05 | Os 3000 inodes da partição acabam |
| Segunda, 14:05 | Novas sessões falham com ENOSPC; a loja responde 503 |
| Segunda, 14:08 | Alerta; o primeiro diagnóstico ("disco com espaço") confunde a equipe |

## Os 5 porquês

1. **Por que a loja deu 503?** A aplicação não conseguia criar o arquivo de sessão.
2. **Por que não conseguia?** Não havia inodes livres na partição, apesar do espaço sobrando.
3. **Por que os inodes acabaram?** Sessões antigas se acumularam por três dias.
4. **Por que se acumularam?** O timer de limpeza estava desativado.
5. **Por que ficou desativado?** Foi pausado na migração sem um item de checklist para religar.

## Fatores que contribuíram

- O monitoramento olhava só o uso de espaço, nunca o de inodes.
- O sintoma era idêntico ao de um incidente anterior, levando a um diagnóstico errado.
- Sessões em arquivo, um por visitante, numa partição pequena.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Alertar uso de inodes, não só de espaço | Detecção |
| Checklist de migração com "religar rotinas pausadas" | Processo |
| Alertar quando um timer crítico estiver desativado | Detecção |
| Avaliar sessões em Redis ou no banco, com expiração automática | Arquitetura |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| `df -h` com espaço livre e `df -i` em 100% | Evidência |
| Milhares de sessões antigas e o timer `limpa-sessoes` desativado | Evidência da causa |
| A loja voltar depois de apagar as sessões | Coincidência: sem o timer, os inodes acabam de novo |

## Para o seu postmortem

Quando o sintoma é igual ao de um incidente anterior, a tentação é aplicar a mesma solução. Confira as hipóteses antes: aqui, `df -h` sozinho dizia que estava tudo bem.
