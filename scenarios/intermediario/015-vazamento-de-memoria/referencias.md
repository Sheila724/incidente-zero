# Para estudar mais

## Zabbix

- [proc.mem e outras chaves do agente](https://www.zabbix.com/documentation/7.0/en/manual/config/items/itemtypes/zabbix_agent): memória por processo
- [Gráficos simples](https://www.zabbix.com/documentation/7.0/en/manual/config/visualization/graphs/simple)
- [Atualizando incidentes (acknowledge)](https://www.zabbix.com/documentation/7.0/en/manual/acknowledgment)
- [Funções de previsão: timeleft() e forecast()](https://www.zabbix.com/documentation/7.0/en/manual/config/triggers/prediction)

## Linux e systemd

- [systemd.resource-control](https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html): `MemoryMax` e outros limites
- [ps](https://man7.org/linux/man-pages/man1/ps.1.html): colunas `rss` e `vsz`

## Para praticar

- Crie no Zabbix do lab um trigger com `timeleft()` que avise quanto tempo falta para a memória chegar a 900 MB.
- Cenários relacionados: **004 — Processo consumindo 100% de CPU** e **014 — Monitoramento cego**.
