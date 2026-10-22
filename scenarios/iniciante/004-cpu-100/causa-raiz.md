# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 15:30 | Deploy do gerador de relatórios com `INTERVAL_SECONDS=0` "para testar" |
| 15:31 | O worker entra em loop e ocupa toda a CPU do servidor |
| 15:32 | A API da loja passa a disputar CPU e fica lenta |
| 15:42 | Alerta de CPU acima de 95% por 10 minutos |

## Os 5 porquês

1. **Por que a loja ficou lenta?** A API não tinha CPU livre para responder.
2. **Por que não tinha CPU?** O report-worker estava usando quase tudo.
3. **Por que o worker usava tanto?** Rodava em loop, sem pausa entre os relatórios.
4. **Por que sem pausa?** O intervalo foi configurado como zero no deploy.
5. **Por que isso chegou em produção?** O valor não tinha validação no código e o serviço não tinha limite de recursos.

## Fatores que contribuíram

- `Restart=always` fez o `kill` parecer resolver, mas o processo voltava em segundos.
- Um serviço de segundo plano rodando no mesmo servidor da loja, sem limite de CPU.
- Mudança de configuração tratada como "só um ajuste", sem revisão.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Validar `INTERVAL_SECONDS` no código (mínimo de 60 s) | Prevenção |
| `CPUQuota` nos serviços de segundo plano | Contenção |
| Alerta de CPU por processo, não só do servidor | Detecção |
| Revisão de mudanças de configuração como se fossem código | Processo |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| `top` e `ps` apontando o `report-worker` com 100% de CPU | Evidência |
| `INTERVAL_SECONDS=0` no arquivo de configuração e o worker gerando relatórios sem pausa | Evidência da causa |
| A CPU cair depois do `kill` | Coincidência: o `Restart=always` traz o processo de volta |

## Para o seu postmortem

Repare na diferença entre **mitigar** (matar o processo, que alivia por segundos), **corrigir** (ajustar a configuração) e **prevenir** (validação e limite de CPU). Um bom postmortem registra os três.
