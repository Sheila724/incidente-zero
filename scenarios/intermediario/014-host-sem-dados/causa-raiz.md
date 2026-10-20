# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| Ontem, 18:00 | Ansible aplica a "padronização" do agente 2 nos servidores web, com o perfil da filial (proxy 10.20.0.50) |
| Ontem, 18:01 | O agente do loja-web-01 passa a recusar o Zabbix Server. Os itens param de receber dados |
| Ontem, 18:01 | Trigger "Agente Zabbix sem comunicação" dispara, com severidade Alta, sem ação de notificação configurada |
| Madrugada | O export de auditoria enche `/var/log`. Nenhum alerta: não há dados para avaliar |
| Hoje | O plantão percebe o host vermelho no painel |

## Os 5 porquês

1. **Por que o Zabbix não tinha dados do host?** O agente recusava as conexões do Zabbix Server.
2. **Por que recusava?** O `Server=` apontava para um proxy desativado.
3. **Por que o arquivo estava errado?** A role do Ansible aplicou o perfil da filial nos servidores da matriz.
4. **Por que ninguém percebeu na hora?** O deploy do Ansible terminou "OK" (o arquivo foi copiado) e não havia teste de ponta a ponta.
5. **Por que o disco encheu sem alerta?** Com o agente mudo, nenhum trigger de disco podia disparar.

## Fatores que contribuíram

- Perfis de configuração parecidos, escolhidos por variável fácil de errar.
- Alerta de "agente sem comunicação" sem notificação para o plantão.
- Export grande gravado em `/var/log`, junto com os logs do sistema.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Teste depois do Ansible: `zabbix_get` a partir do servidor, para cada host | Detecção |
| Notificar o plantão quando um agente fica sem comunicação | Detecção |
| Dashboard de hosts sem dados ("falha silenciosa") | Detecção |
| Perfis do Ansible com validação (o IP do servidor precisa existir) | Prevenção |
| Exports em partição ou armazenamento próprios | Prevenção |

## Para o seu postmortem

Um monitoramento que não alerta sobre a própria cegueira dá uma falsa sensação de segurança. Quando um host fica sem dados, a pergunta certa não é só "por que o agente parou?", mas também **"o que aconteceu enquanto não estávamos olhando?"**
