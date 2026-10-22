# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Mês passado | Tuning de memória (CHG-311): `max_connections` reduzido de 151 para 30 |
| Ontem, 22:00 | BI publica a v2 do relatório: 5 conexões em paralelo e espera de até 10 min pelo FTP |
| Hoje, ~10:20 | O FTP do financeiro fica lento; cada execução passa a segurar as conexões por 10 minutos |
| ~10:26 | 6 execuções simultâneas × 5 conexões = 30. O banco enche |
| 10:32 | O monitoramento alerta: loja com 503 |

## Os 5 porquês

1. **Por que a loja caiu?** O banco recusava conexões novas (erro 1040).
2. **Por que estava cheio?** O usuário `relatorio` ocupava todas as 30 conexões, paradas em `Sleep`.
3. **Por que o relatório segurava conexões?** A v2 abre uma conexão por consulta e não fecha antes de esperar o FTP.
4. **Por que isso afetou a loja?** Loja e relatório dividem o mesmo limite, sem limite por usuário.
5. **Por que ninguém percebeu antes?** Com o FTP rápido, cada execução terminava em segundos; o vazamento só apareceu quando a espera aumentou.

## Fatores que contribuíram

- `max_connections` reduzido sem revisar quem consome conexões.
- Mudança de código do BI sem revisão do time responsável pelo banco.
- Nenhum alerta para `Threads_connected` perto do limite.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Fechar conexões com `with` e revisar código que acessa o banco | Prevenção |
| `MAX_USER_CONNECTIONS` para usuários de relatório e integrações | Prevenção |
| Alerta quando `Threads_connected` passar de 80% de `max_connections` | Detecção |
| Relatórios pesados numa réplica de leitura, não no banco da loja | Prevenção |
| `wait_timeout` menor para usuários de batch | Prevenção |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| `SHOW PROCESSLIST` com todas as conexões do usuário `relatorio` em `Sleep` | Evidência |
| O script do relatório abrindo uma conexão por categoria e dormindo sem fechar | Evidência da causa |
| A loja voltar depois do `restart` do MariaDB | Coincidência: o cron enche o banco de novo em poucos minutos |

## Para o seu postmortem

O relatório não "derrubou o banco": ele usou um recurso compartilhado sem limite, e a loja ficou sem a sua parte. Sempre que dois sistemas dividem algo finito (conexões, disco, CPU), pergunte: **o que impede um deles de pegar tudo?**
