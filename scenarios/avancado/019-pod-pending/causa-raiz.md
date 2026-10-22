# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 10:02 | Primeiro deploy do `loja-checkout`: a API aceita os manifestos |
| 10:02 | Os 2 pods ficam em `Pending`: `0/1 nodes are available: 1 Insufficient cpu` |
| 10:02 | O Service do checkout fica sem endpoints e o nginx responde 502 em `/checkout/` |
| 10:07 | O monitoramento abre o alerta |

## Os 5 porquês

1. **Por que o checkout respondia 502?** Nenhum pod do `loja-checkout` estava rodando.
2. **Por que nenhum pod rodava?** O scheduler não encontrou nó para eles (`Pending`).
3. **Por que não encontrou?** Cada pod reservava 250 núcleos de CPU.
4. **Por que 250 núcleos?** O manifesto trazia `cpu: "250"`, sem o `m` de milicores.
5. **Por que passou?** A API só valida o formato, não se o valor cabe em algum nó, e o pipeline considerou "aplicado" como "no ar".

## Fatores que contribuíram

- Unidades parecidas com significados mil vezes diferentes.
- Pipeline sem `kubectl rollout status` depois do apply.
- Primeiro deploy de um serviço novo, sem revisão de quem opera o cluster.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| `kubectl rollout status` (com timeout) no pipeline: deploy só termina com os pods prontos | Detecção |
| `LimitRange` no namespace, com máximo por container | Prevenção |
| Validação dos manifestos no CI (ex.: políticas com Kyverno ou conftest) | Prevenção |
| Alerta para pods em `Pending` por mais de alguns minutos | Detecção |

## Evidência × coincidência

| O que foi observado | Peso |
|---|---|
| Evento `FailedScheduling: Insufficient cpu` | Evidência do caminho do erro |
| `requests.cpu: "250"` no Deployment e poucos núcleos em `Allocatable` no nó | Evidência da causa |
| A API ter aceitado o deploy sem erro | Não é sinal de sucesso: aceitar não é agendar |
| O pod subir depois de remover os requests | Coincidência: some o sintoma, fica o risco |

## Para o seu postmortem

"Aplicado" não é "no ar". Entre o `kubectl apply` e o pod rodando ainda existem o scheduler, o download da imagem e as probes, e cada etapa tem o seu jeito de falhar.
