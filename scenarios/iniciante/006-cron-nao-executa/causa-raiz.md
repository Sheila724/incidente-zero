# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| Segunda, 09:00 | "Padronização": o agendamento ganha o nome `backup-loja.cron` |
| Segunda, 09:00 | Na mesma mudança, o log passa a ter a data no nome com `$(date +%F)` |
| Segunda → quinta | O cron ignora o arquivo; nenhum backup é feito |
| Quinta, 09:15 | Alerta: backup mais recente com 72 horas |

## Os 5 porquês

1. **Por que não há backup há 3 dias?** O agendamento não executou.
2. **Por que não executou?** O cron ignorava o arquivo por causa do ponto no nome.
3. **Por que, depois de renomear, ainda falhava?** O `%` sem escape cortava o comando.
4. **Por que ninguém viu os erros?** O cron não avisa sobre arquivos ignorados, e a saída de erro ia para um e-mail que não existe.
5. **Por que demorou 3 dias para alertar?** O alerta só olhava a idade do backup, com limite de 24 horas e verificação diária.

## Fatores que contribuíram

- Mudança "cosmética" em arquivos de agendamento, sem teste.
- Duas falhas sobrepostas: corrigir a primeira não resolvia o incidente.
- Saída do cron sem destino (sem `MAILTO` e sem servidor de e-mail).

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Lógica de datas dentro do script, não na linha do cron | Prevenção |
| Lint dos arquivos do `/etc/cron.d` (nomes e `%`) no pipeline | Prevenção |
| Alerta de "backup não rodou" em minutos, não em dias (heartbeat) | Detecção |
| Teste de restauração periódico | Validação |

## Para o seu postmortem

Falhas silenciosas são as mais caras: tudo parecia normal por três dias. Pergunte-se sempre: **se isso parar de funcionar, quem fica sabendo, e em quanto tempo?**
