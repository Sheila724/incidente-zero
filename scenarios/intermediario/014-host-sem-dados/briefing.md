# 🚨 Incidente: monitoramento cego

**Severidade:** alta
**Servidor:** `loja-web-01`
**Origem:** Zabbix

> **[ZABBIX] PROBLEMA: Agente Zabbix sem comunicação em loja-web-01**
> Severidade: Alta. Nenhum dado recebido do host.

A loja parece estar funcionando, e ninguém reclamou. Mas o Zabbix não recebe nada desse servidor: CPU, disco, memória, nada.

No canal de infraestrutura, um aviso de ontem: *"Rodamos a padronização dos agentes Zabbix com o Ansible em todos os servidores web. Agente 2 atualizado e configuração unificada."*

## Sua missão

1. Fazer o Zabbix voltar a receber dados do loja-web-01.
2. Descobrir por que o agente parou de responder.
3. Conferir se algo aconteceu **enquanto o monitoramento estava cego**.

O Zabbix deste ambiente está no botão **Zabbix**, no topo da tela (usuário `Admin`, senha `zabbix`). O Zabbix Server roda nesta mesma máquina, no endereço `127.0.0.1`.

Você está logado como `aluno`, com acesso `sudo`.
