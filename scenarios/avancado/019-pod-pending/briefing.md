# 🚨 Incidente: checkout novo não sobe

**Severidade:** alta
**Servidor:** `loja-k8s-01` (Kubernetes, nó único)
**Origem:** monitoramento

> **[ALERTA] Checkout da Loja Exemplo retornando HTTP 502**
> Verificação externa falhou 5 vezes seguidas em http://loja-k8s-01/checkout/

O time de produto acabou de fazer o **primeiro deploy** do novo serviço de checkout: o Deployment `loja-checkout`, com 2 réplicas, exposto na porta 30081. O nginx do servidor manda `/checkout/` para ele e o resto do site para a `loja-web`, que continua funcionando.

O pipeline aplicou os manifestos sem erro. "O Kubernetes aceitou, então está no ar", disse quem fez o deploy.

## Sua missão

1. Descobrir por que o checkout não está atendendo.
2. Colocar o checkout no ar, com as 2 réplicas prontas.
3. Fazer isso sem prejudicar a loja principal e mantendo a reserva de recursos do checkout.

Você está logado como `aluno`. O `kubectl` já está configurado e aponta para o namespace `loja`. A documentação do serviço está em `/opt/loja/checkout/`.
