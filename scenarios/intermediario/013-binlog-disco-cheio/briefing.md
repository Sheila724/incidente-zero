# 🚨 Incidente: pedidos não estão sendo gravados

**Severidade:** crítica
**Servidor:** `loja-db-01`
**Horário do alerta:** 06:48

> **[ALERTA] Checkout de http://loja-db-01/ falhando**
> Nenhum pedido gravado nos últimos 20 minutos. Catálogo funcionando normalmente.

Os clientes navegam pela loja, mas nenhuma compra é concluída. O MariaDB está "ativo" segundo o systemd.

Contexto do plantão anterior: *"Semana passada ligamos o binlog para testar a réplica loja-db-02 (CHG-340). O teste acabou e a réplica foi desligada na sexta. Ah, e esta madrugada rodou a importação de eventos do marketing."*

## Sua missão

1. Fazer o checkout voltar a gravar pedidos.
2. Liberar espaço **sem corromper** o banco nem o histórico do binlog.
3. Garantir que isso não aconteça de novo.

Você está logado como `aluno`, com acesso `sudo`. O root do MariaDB entra pelo socket local: `sudo mariadb`.
