# 🚨 Incidente: deploy da 2.6 travado

**Severidade:** alta
**Servidor:** `loja-k8s-01` (Kubernetes, nó único)
**Origem:** pipeline de deploy

> **[PIPELINE] FALHOU: deploy loja-web 2.6**
> `kubectl rollout status deploy/loja-web`: deployment "loja-web" exceeded its progress deadline

A versão **2.6** da Loja Exemplo traz a página da promoção que o marketing anuncia às 10:00. O pipeline publicou a imagem e aplicou o deploy, mas o rollout nunca terminou.

Os clientes ainda conseguem comprar, na versão antiga. Por enquanto, ninguém sabe dizer em que estado o cluster ficou.

## Sua missão

1. Descobrir por que o rollout não avança.
2. Explicar por que a loja não caiu.
3. Colocar a **2.6** no ar, com todas as réplicas prontas. Voltar para a 2.5 não resolve o problema do marketing.

Você está logado como `aluno`. O `kubectl` já está configurado e aponta para o namespace `loja`. Os logs do pipeline ficam em `/var/log/pipeline/`.
