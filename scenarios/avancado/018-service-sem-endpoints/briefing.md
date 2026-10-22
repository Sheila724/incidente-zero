# 🚨 Incidente: loja fora do ar, pods "rodando"

**Severidade:** crítica
**Servidor:** `loja-k8s-01` (Kubernetes, nó único)
**Origem:** monitoramento

> **[ALERTA] Loja Exemplo retornando HTTP 502**
> Verificação externa falhou 5 vezes seguidas em http://loja-k8s-01/

A Loja Exemplo roda no Kubernetes deste servidor: um **Deployment** `loja-web` com 2 réplicas, exposto pelo **Service** `loja-web` na porta 30080. O nginx do servidor recebe os clientes na porta 80 e repassa para esse Service.

O Deployment usa a estratégia **Recreate** (a subida da loja faz migração do banco, então duas versões não podem rodar juntas).

Há 10 minutos, o time de plataforma aplicou a "padronização das health checks" em todos os serviços do cluster. Quem olhou o painel diz que "está tudo Running".

## Sua missão

1. Descobrir por que a loja responde 502 com os pods de pé.
2. Colocar a loja no ar de novo, com as 2 réplicas prontas.
3. Manter a verificação de prontidão funcionando: ela protege os clientes durante deploys.

Você está logado como `aluno`. O `kubectl` já está configurado e aponta para o namespace `loja`.
