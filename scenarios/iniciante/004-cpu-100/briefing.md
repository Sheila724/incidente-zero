# 🚨 Incidente: API da loja lenta

**Severidade:** alta
**Servidor:** `loja-web-01`
**Horário do alerta:** 15:42

> **[ALERTA] CPU acima de 95% há 10 minutos em loja-web-01**
> p95 de resposta da API passou de 4 segundos.

O site está no ar, mas as páginas demoram para abrir e o checkout está dando timeout para parte dos clientes.

No canal do time, alguém comenta que hoje à tarde saiu uma versão nova do **gerador de relatórios de vendas**, e que "foi só um ajuste de configuração".

## Sua missão

1. Descobrir o que está consumindo a CPU.
2. Devolver a performance da loja **sem desligar o gerador de relatórios**: o financeiro precisa deles.
3. Garantir que o problema não volte no próximo restart.

Você está logado como `aluno`, com acesso `sudo`.
