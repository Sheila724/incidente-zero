# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| Semana passada | CHG-340: binlog ligado para testar a réplica loja-db-02, com `binlog_expire_logs_seconds = 0` |
| Sexta-feira | Teste concluído, réplica desligada. O binlog continua ligado |
| Madrugada | Importação de eventos do marketing: grava e limpa uma tabela de staging centenas de vezes |
| 06:28 | A partição `/var/lib/mysql` chega a 100%. O MariaDB passa a esperar espaço; nenhum pedido é gravado |
| 06:48 | O monitoramento alerta |

## Os 5 porquês

1. **Por que os pedidos não eram gravados?** O MariaDB estava parado esperando espaço em disco.
2. **Por que o disco encheu?** Os binlogs ocupavam quase toda a partição do banco.
3. **Por que tanto binlog?** A importação gerou um volume grande de alterações, e nada era apagado.
4. **Por que nada era apagado?** A expiração estava desligada para a réplica de teste, que já não existia.
5. **Por que ninguém percebeu?** Não havia alerta de uso da partição do banco, só do disco raiz.

## Fatores que contribuíram

- Mudança temporária (CHG-340) sem passo de reversão.
- Binlog e dados na mesma partição.
- Job de importação pesado rodando direto no banco de produção.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Expiração do binlog sempre configurada (ou binlog desligado se não há réplica) | Prevenção |
| Alerta de uso da partição do banco em 80% e 90% | Detecção |
| Toda mudança temporária com data e passo de reversão | Processo |
| Importações pesadas em tabela fora do binlog (`SET sql_log_bin = 0` quando aplicável) ou em outro servidor | Prevenção |
| Runbook "disco cheio no banco": liberar espaço no sistema de arquivos antes de qualquer comando SQL | Processo |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| Log do MariaDB com `Disk is full writing ... Waiting for someone to free space` e consultas presas nesse estado | Evidência |
| Os `mysql-bin.*` somando a maior parte da partição, com expiração desligada | Evidência da causa |
| O banco voltar depois de mover os binlogs | Mitigação: sem `PURGE` e expiração, o índice fica inconsistente e o disco enche de novo |

## Para o seu postmortem

O instinto num banco travado é reiniciar. Aqui, isso podia transformar um incidente de gravação em um banco corrompido. Antes de agir, entenda **o que o servidor está esperando**: o log dizia, com todas as letras, *"Waiting for someone to free space"*.
