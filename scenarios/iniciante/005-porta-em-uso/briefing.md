# 🚨 Incidente: site exibindo arquivos do servidor

**Severidade:** crítica
**Servidor:** `loja-web-01`
**Horário do alerta:** 08:05

> **[CLIENTE] "Entrei no site de vocês e apareceu uma lista de pastas: bin, etc, home…"**
> Chamado aberto pelo atendimento às 08:02.

O servidor foi reiniciado de madrugada para uma atualização do sistema. Desde então, quem acessa o endereço da loja vê uma página estranha, com uma lista de arquivos, em vez do site.

O monitoramento não alertou, porque o site continua respondendo HTTP 200.

## Sua missão

1. Tirar do ar o que está sendo exposto, **o quanto antes**.
2. Colocar a Loja Exemplo de volta no lugar.
3. Garantir que isso não aconteça no próximo reboot.

Você está logado como `aluno`, com acesso `sudo`.
