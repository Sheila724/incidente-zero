# 🚨 Incidente: 502 depois do deploy

**Severidade:** crítica
**Servidor:** `loja-web-01`
**Horário do alerta:** 19:04

> **[ALERTA] http://loja-web-01/ retornando HTTP 502 Bad Gateway**
> Disponibilidade da loja em 0% nos últimos 3 minutos.

O deploy da versão **2.4.0** (página de promoções da Black Friday) terminou às 19:01. O pipeline mostra tudo verde e o serviço `app` está ativo.

O marketing já disparou o e-mail da promoção para 400 mil clientes às 19:00.

## Sua missão

1. Colocar a loja de volta no ar **sem desfazer o deploy**: a promoção precisa estar no ar.
2. Garantir que o site **e** a API usada pelo app mobile estejam funcionando.
3. Entender o que o deploy mudou e por que o pipeline não pegou.

Você está logado como `aluno`, com acesso `sudo`.
