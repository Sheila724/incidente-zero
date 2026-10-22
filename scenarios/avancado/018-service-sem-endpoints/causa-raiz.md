# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 14:00 | O time de plataforma aplica a "padronização das health checks": todas as readiness probes passam a chamar `/healthz` |
| 14:00 | O Deployment `loja-web` (estratégia Recreate) derruba os 2 pods e cria 2 novos |
| 14:00 | Os pods novos sobem, mas a probe recebe 404; nenhum fica pronto |
| 14:01 | O Service fica sem endpoints e o nginx passa a responder 502 |

## Os 5 porquês

1. **Por que a loja respondia 502?** O Service `loja-web` não tinha endpoints.
2. **Por que sem endpoints?** Nenhum pod passava na readiness probe.
3. **Por que a probe falhava?** Ela chamava `/healthz`, e a loja só atende `/health`.
4. **Por que a probe mudou?** Uma padronização foi aplicada a todos os serviços de uma vez, sem conferir o que cada aplicação atende.
5. **Por que o impacto foi total?** Com a estratégia Recreate, todos os pods antigos saíram antes de os novos ficarem prontos. Com rolling update, os antigos teriam segurado o tráfego.

## Fatores que contribuíram

- "Está tudo Running" foi lido como "está tudo bem".
- A mudança veio de outro time e não passou pelo deploy da loja.
- Recreate não tem rede de proteção: se a versão nova não fica pronta, não sobra nada.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Padronizações aplicadas serviço a serviço, conferindo o endpoint antes | Prevenção |
| A aplicação passa a atender `/healthz` antes de a probe mudar | Prevenção |
| Teste no pipeline que sobe o pod e valida as probes antes do deploy | Prevenção |
| Alerta para Service sem endpoints e para Deployment com réplicas prontas abaixo do desejado | Detecção |

## Evidência × coincidência

| O que foi observado | Peso |
|---|---|
| `Endpoints:` vazio no Service | Evidência do caminho do erro |
| Eventos `Readiness probe failed: ... statuscode: 404` | Evidência |
| `/health` responde 200 e `/healthz` responde 404 dentro do pod | Evidência da causa |
| Pods `Running` sem reinícios | Não é sinal de saúde: readiness não reinicia o pod |

## Para o seu postmortem

Num cluster, "está de pé" e "está recebendo tráfego" são perguntas diferentes. Siga o caminho do pedido (nginx → Service → endpoints → pod) e veja onde ele para.
