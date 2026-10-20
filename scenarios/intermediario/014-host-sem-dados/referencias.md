# Para estudar mais

## Zabbix

- [Agente 2: parâmetros de configuração](https://www.zabbix.com/documentation/7.0/en/manual/appendix/config/zabbix_agent2): `Server`, `ServerActive`, `Hostname`
- [Checagens passivas e ativas](https://www.zabbix.com/documentation/7.0/en/manual/appendix/items/activepassive)
- [zabbix_get](https://www.zabbix.com/documentation/7.0/en/manual/concepts/get)
- [Funções de histórico, como nodata()](https://www.zabbix.com/documentation/7.0/en/manual/appendix/functions/history): detectar falhas silenciosas
- [API do Zabbix](https://www.zabbix.com/documentation/7.0/en/manual/api)

## Para praticar

- No Zabbix do lab, crie uma ação que avise quando um agente ficar sem comunicação.
- Use a API: `zbx-api problem.get '{"output":["name","severity"]}'`.
- Projeto relacionado: o [NOC Horizon](https://github.com/Sheila724/noc-horizon) destaca hosts "sem dados" justamente por causa desse tipo de incidente.
- Cenários relacionados: **001 — Disco cheio em /var/log** e **015**.
