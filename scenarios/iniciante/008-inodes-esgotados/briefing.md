# 🚨 Incidente: loja fora do ar com disco "cheio"

**Severidade:** alta
**Servidor:** `loja-web-01`
**Horário do alerta:** 14:08

> **[ALERTA] HTTP 503 em http://loja-web-01/**
> Log da aplicação: `OSError: [Errno 28] No space left on device`

Parece o mesmo incidente da semana passada, de disco cheio. Só que quem atendeu primeiro jura que **o disco está com mais de 90% livre**.

Na sexta-feira houve uma migração de servidor, e algumas rotinas automáticas foram pausadas "por segurança" durante a janela.

## Sua missão

1. Entender como pode faltar espaço com o disco quase vazio.
2. Colocar a loja de volta no ar **sem deslogar os clientes que estão comprando agora**.
3. Garantir que o problema não volte.

Você está logado como `aluno`, com acesso `sudo`.
