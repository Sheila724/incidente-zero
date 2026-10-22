# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Ontem, 18:40 | A ConfigMap `loja-config` é editada: `DATABASE_URL` vira `DB_URL` ("padronização dos nomes") |
| Ontem, 18:40 em diante | Nada acontece: os pods em execução já tinham a variável carregada |
| 03:00 | A manutenção do cluster recria todos os pods |
| 03:01 | Os pods novos sobem sem `DATABASE_URL`, encerram e entram em `CrashLoopBackOff` |
| 03:02 | O Service fica sem pods prontos e o nginx passa a responder 502 |

## Os 5 porquês

1. **Por que a loja caiu?** Nenhum pod da `loja-web` ficava pronto.
2. **Por que os pods não ficavam prontos?** A aplicação encerrava na subida por falta de `DATABASE_URL`.
3. **Por que a variável sumiu?** A chave da ConfigMap foi renomeada sem mudar a aplicação.
4. **Por que ninguém percebeu na hora?** Variáveis de ambiente só são lidas quando o container inicia. O erro ficou "dormindo" até o próximo reinício.
5. **Por que a mudança passou?** A ConfigMap foi editada direto no cluster, sem revisão e sem um teste que subisse a aplicação com ela.

## Fatores que contribuíram

- Mudança de configuração fora do fluxo de deploy (sem revisão, sem rollout).
- O efeito apareceu horas depois, em outro evento (a manutenção), o que confunde a investigação.
- A aplicação não valida a configuração em tempo de deploy, só em tempo de subida.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| ConfigMaps versionadas no Git e aplicadas pelo pipeline, com revisão | Prevenção |
| Toda mudança de ConfigMap dispara um rollout (ex.: hash da configuração como anotação do pod) | Prevenção |
| Renomear com transição: manter o nome antigo até a aplicação usar o novo | Prevenção |
| Alerta para pods com reinícios em sequência (`kube_pod_container_status_restarts_total`) | Detecção |

## Evidência × coincidência

| O que foi observado | Peso |
|---|---|
| `kubectl logs` e o `Last State` do pod com "DATABASE_URL não definida" | Evidência |
| A ConfigMap sem `DATABASE_URL` e a anotação da mudança de ontem | Evidência da causa |
| A loja ter caído logo depois da manutenção | Coincidência: a manutenção só revelou um erro que já estava lá |
| Apagar um pod e o novo também falhar | Evidência de que o problema não está no pod, e sim no que ele herda |

## Para o seu postmortem

A pergunta "o que mudou?" precisa olhar para trás no tempo, não só para os últimos minutos. Em Kubernetes, uma mudança de configuração pode esperar horas até o próximo reinício para aparecer.
