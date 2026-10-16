# 🚨 Incidente: Loja Exemplo com acesso negado

**Severidade:** alta
**Servidor:** `loja-web-01`
**Horário do alerta:** 10:32

> **[ALERTA] HTTP 403 em http://loja-web-01/**
> A página inicial passou a retornar "403 Forbidden".

Clientes relatam que a loja mostra **"403 Forbidden"** e que as imagens e a página de produtos não carregam.

O time de segurança comenta no canal que hoje cedo entrou em produção uma melhoria pedida pela auditoria (ticket **SEC-142**) para "proteger os arquivos do site". Ninguém mexeu no nginx.

## Sua missão

1. Colocar a loja de volta no ar.
2. Garantir que o problema **não volte sozinho**.
3. Fazer isso **sem desfazer a proteção pedida pela auditoria**, se possível.

Você está logado como `aluno`, com acesso `sudo`.
