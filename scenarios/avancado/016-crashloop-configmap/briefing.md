# 🚨 Incidente: loja fora do ar após manutenção

**Severidade:** crítica
**Servidor:** `loja-k8s-01` (Kubernetes, nó único)
**Origem:** monitoramento

> **[ALERTA] Loja Exemplo retornando HTTP 502**
> Verificação externa falhou 5 vezes seguidas em http://loja-k8s-01/

A Loja Exemplo roda no Kubernetes deste servidor: um **Deployment** `loja-web` com 2 réplicas, exposto por um **Service** na porta 30080. O nginx do servidor recebe os clientes na porta 80 e repassa para esse Service.

Às 03:00 houve a manutenção programada do cluster, e todos os pods foram recriados. Ninguém fez deploy de versão nova desde ontem.

## Sua missão

1. Descobrir por que os pods da loja não ficam de pé.
2. Colocar a loja no ar de novo, com as 2 réplicas prontas.
3. Entender o que a manutenção teve a ver com isso.

Você está logado como `aluno`. O `kubectl` já está configurado e aponta para o namespace `loja`.
