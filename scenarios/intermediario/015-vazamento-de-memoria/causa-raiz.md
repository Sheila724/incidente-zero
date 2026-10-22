# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 07:00 | Deploy da 2.5.0 com o cache de sessões; `SESSION_CACHE_MAX=0` no arquivo de ambiente |
| 07:00 em diante | Cada visita guarda ~50 KB que nunca saem da memória |
| 09:40 | A memória do processo passa de 250 MB e o Zabbix abre o incidente |
| Sem correção | Em ~20 minutos, chegaria a 900 MB e o systemd encerraria a loja, em ciclos |

## Os 5 porquês

1. **Por que a memória subia sem parar?** O cache de sessões guardava todas as sessões, sem limite.
2. **Por que sem limite?** O deploy usou `SESSION_CACHE_MAX=0`, que significa "ilimitado".
3. **Por que esse valor?** O padrão do parâmetro novo foi o mais "seguro" para o desenvolvedor (não perder sessões), não para o servidor.
4. **Por que não apareceu nos testes?** Em homologação, o tráfego é pequeno e a memória cresce devagar.
5. **Por que foi detectado a tempo?** Havia um item de memória do processo com trigger antes do limite do systemd.

## Fatores que contribuíram

- Parâmetro novo com padrão perigoso (0 = ilimitado).
- Teste de carga ausente no pipeline.
- Uma rampa lenta não chama atenção sem um gráfico.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Padrões seguros para parâmetros novos (sempre com limite) | Prevenção |
| Teste de carga de alguns minutos antes de liberar a versão | Prevenção |
| Trigger de tendência, ex.: `trendavg` ou previsão com `timeleft()` | Detecção |
| Dashboard de memória por versão após cada deploy | Detecção |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| O gráfico do Zabbix com uma rampa contínua desde o horário do deploy e o `/health` mostrando as sessões crescendo | Evidência |
| `SESSION_CACHE_MAX=0` (sem limite) no arquivo de ambiente da 2.5.0 | Evidência da causa |
| A memória cair depois do `restart` da loja | Coincidência: a rampa recomeça no mesmo ritmo |

## Para o seu postmortem

Reiniciar é uma ótima mitigação e uma péssima correção. Se depois do restart o gráfico volta a desenhar a mesma rampa, o problema ainda está lá. E um incidente sem registro no Zabbix obriga o próximo plantão a descobrir tudo de novo.
