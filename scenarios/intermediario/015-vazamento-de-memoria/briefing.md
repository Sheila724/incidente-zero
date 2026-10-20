# 🚨 Incidente: memória da loja subindo

**Severidade:** média
**Servidor:** `loja-web-01`
**Origem:** Zabbix

> **[ZABBIX] PROBLEMA: Loja Exemplo usando muita memória em loja-web-01**
> Severidade: Média. Memória do processo acima de 250 MB.

A loja está no ar e respondendo. Por enquanto. O processo tem um limite de 900 MB: se chegar lá, o systemd o encerra, e a loja cai no meio das compras.

O deploy da versão **2.5.0** foi feito hoje cedo, com a novidade do "carrinho até 10x mais rápido".

## Sua missão

1. Entender, pelo Zabbix, o comportamento da memória e desde quando ele começou.
2. Estabilizar a loja **sem desfazer o deploy** e sem tirar os clientes do ar.
3. Registrar no Zabbix o que foi feito, para o próximo plantão.

O Zabbix deste ambiente está no botão **Zabbix**, no topo da tela (usuário `Admin`, senha `zabbix`).

Você está logado como `aluno`, com acesso `sudo`.
