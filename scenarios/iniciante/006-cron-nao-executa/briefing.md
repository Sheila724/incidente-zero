# ⚠️ Incidente: backups parados há 3 dias

**Severidade:** média
**Servidor:** `loja-db-01`
**Horário do alerta:** 09:15

> **[ALERTA] Backup mais recente de /srv/loja/data tem 72 horas**
> Limite configurado: 24 horas.

Nada está fora do ar, mas a loja está sem backup há três dias. Se o servidor falhar agora, perdemos três dias de pedidos.

Na segunda-feira houve uma "padronização" dos arquivos de agendamento do servidor. O time de banco também pediu para o backup rodar **a cada 2 minutos** enquanto a rotina não for validada.

## Sua missão

1. Fazer o backup voltar a rodar sozinho, pelo agendamento.
2. Confirmar que o backup gerado é **válido**.
3. Entender por que ele parou sem ninguém perceber.

Você está logado como `aluno`, com acesso `sudo`.
