# 🚨 Incidente: pipeline de deploy bloqueado

**Severidade:** alta
**Servidor:** `app-01`
**Horário do alerta:** 11:20

> **[PIPELINE] deploy-producao #482 falhou**
> `deploy@app-01: Permission denied (publickey).`

Desde ontem à tarde nenhum deploy chega em produção, e há uma correção urgente esperando na fila.

O que se sabe:

- Ontem, a chave do pipeline foi **trocada** como parte da rotação trimestral de credenciais.
- Também ontem, alguém do suporte "resolveu um problema de permissão" no home do usuário `deploy` para conseguir subir um arquivo.

A chave privada do pipeline está em `~/ci/deploy_key` para você testar o mesmo acesso.

## Sua missão

1. Fazer o pipeline voltar a entrar no servidor como `deploy`.
2. Fazer isso **sem enfraquecer o SSH** (nada de liberar senha ou desligar verificações).
3. Desfazer o estrago do "ajuste de permissão".

Você está logado como `aluno`, com acesso `sudo`.
