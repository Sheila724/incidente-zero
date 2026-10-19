# 🚨 Incidente: checkout fora do ar

**Severidade:** crítica
**Servidor:** `loja-web-01`
**Horário do alerta:** 02:14

> **[ALERTA] http://loja-web-01/checkout retornando HTTP 502**
> 100% das tentativas de compra falhando há 9 minutos.

A home e o catálogo funcionam, mas ninguém consegue finalizar uma compra.

O time de pagamentos avisa no canal: *"Migramos a API de pagamentos do pag-api-01 para o pag-api-02 às 02:00, conforme a janela aprovada. O DNS interno já foi atualizado e nossos testes passaram. Do nosso lado está tudo certo."*

## Sua missão

1. Fazer o checkout voltar a funcionar.
2. Explicar por que a loja não acompanhou a migração, se o DNS já estava certo.
3. Deixar o servidor pronto para a próxima migração, sem depender de ajuste manual.

Você está logado como `aluno`, com acesso `sudo`.
