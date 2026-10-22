# 🚨 Incidente: loja fora do ar de madrugada

**Severidade:** crítica
**Servidor:** `loja-k8s-01` (Kubernetes, nó único)
**Origem:** monitoramento
**Horário:** 02:10, você é o plantão

> **[ALERTA] Loja Exemplo retornando HTTP 502**
> Verificação externa falhou 5 vezes seguidas em http://loja-k8s-01/

A Loja Exemplo roda no Kubernetes deste servidor:

```
cliente → nginx :80 → Service loja-web → pods loja-web (2) → Service loja-db → pod loja-db
```

Ontem à tarde, o time de dados trouxe o banco (`loja-db`) para dentro do cluster. "Testamos tudo em homologação e funcionou", disseram. Não houve deploy da loja desde então.

O monitoramento mostra que a loja já caiu e voltou algumas vezes nas últimas horas, sem ninguém mexer.

## Sua missão

1. Colocar a loja de volta no ar o quanto antes.
2. Descobrir por que ela cai, volta e cai de novo.
3. Corrigir a causa sem desligar as automações do cluster que funcionam.

Você está logado como `aluno`. O `kubectl` já está configurado e aponta para o namespace `loja`. As anotações do time de infraestrutura ficam em `/opt/loja/infra/`.
