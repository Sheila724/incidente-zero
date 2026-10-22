# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| D-1, 22:40 | `LOG_LEVEL` alterado para `DEBUG` para investigar lentidão no checkout |
| D-1, 23:00 | Investigação encerrada; a configuração não foi revertida |
| D-1 → D | `debug.log` cresce sem parar, sem rotação |
| D, 03:10 | `/var/log` atinge 100% |
| D, 03:12 | A aplicação falha ao gravar e responde 503; o alerta dispara |

## Os 5 porquês

1. **Por que o site caiu?** A aplicação não conseguia gravar no disco e respondia 503.
2. **Por que não conseguia gravar?** `/var/log` estava 100% cheio.
3. **Por que encheu?** O `debug.log` cresceu centenas de MB em poucas horas.
4. **Por que cresceu tanto?** O nível DEBUG ficou ligado em produção.
5. **Por que ficou ligado?** Era uma mudança "temporária", sem registro, sem prazo e sem alguém responsável por desfazer.

## Fatores que contribuíram

- Nenhuma rotação de log configurada para a aplicação.
- Nenhum alerta de disco **antes** de chegar a 100%. O primeiro sinal foi o site fora do ar.
- A aplicação trata falha de log como falha da requisição. Logar não deveria derrubar o serviço.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Configurar logrotate com `maxsize` para `/var/log/app` | Prevenção |
| Alertar com disco em 80% (aviso) e 90% (crítico) | Detecção |
| Registrar mudanças temporárias em ticket, com prazo para reverter | Processo |
| Fazer a aplicação continuar atendendo mesmo sem conseguir logar | Resiliência |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| `journalctl -u app` com `[Errno 28] No space left on device`, `df` em 100% e `debug.log` ocupando a partição | Evidência |
| `LOG_LEVEL=DEBUG` no arquivo **e** no `/health` da aplicação em execução | Evidência |
| O site voltar depois do `truncate` | Coincidência: só libera espaço; com DEBUG ligado, o disco enche de novo |

## Para o seu postmortem

Um bom postmortem é **sem culpados**: o foco é o sistema e o processo, não quem esqueceu de reverter. Pergunte "o que tornou esse erro fácil de cometer e difícil de perceber?".
