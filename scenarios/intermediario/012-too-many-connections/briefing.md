# 🚨 Incidente: banco recusando conexões

**Severidade:** crítica
**Servidor:** `loja-db-01`
**Horário do alerta:** 10:32

> **[ALERTA] http://loja-db-01/ retornando HTTP 503**
> Loja e checkout indisponíveis. Erros de banco de dados na aplicação.

A loja inteira está fora do ar. O MariaDB está rodando, o disco está livre e a CPU está tranquila.

No canal do time de BI, uma mensagem de ontem à noite: *"Publicamos a v2 do relatório de vendas: agora as consultas rodam em paralelo, ficou muito mais rápido! 🚀"*

## Sua missão

1. Colocar a loja de volta no ar.
2. Descobrir quem está consumindo as conexões do banco.
3. Corrigir a causa **sem desligar o relatório**: o financeiro depende dele.

Você está logado como `aluno`, com acesso `sudo`. O root do MariaDB entra pelo socket local: `sudo mariadb`.
