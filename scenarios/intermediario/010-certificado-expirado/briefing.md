# 🚨 Incidente: certificado expirado

**Severidade:** alta
**Servidor:** `loja-web-01`
**Horário do alerta:** 08:02

> **[ALERTA] Certificado TLS de https://loja.local expirado**
> A verificação de TLS falha desde ontem às 18h.

O suporte está recebendo reclamações desde cedo: clientes veem *"Sua conexão não é particular"* no navegador, e o app da loja não consegue mais falar com o site. As vendas caíram 80%.

Contexto do time de infraestrutura: *"Esse certificado é renovado sozinho, nunca precisamos mexer nele. No mês passado migramos o servidor para o template novo, mas copiamos tudo do antigo."*

## Sua missão

1. Colocar o HTTPS de volta no ar com um certificado válido.
2. Descobrir por que a renovação automática não aconteceu.
3. Garantir que a próxima renovação funcione sozinha, do começo ao fim.

Para testar como um cliente (sem DNS), use:

```bash
curl -v --resolve loja.local:443:127.0.0.1 https://loja.local/
```

Você está logado como `aluno`, com acesso `sudo`.
