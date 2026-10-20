# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja o problema no Zabbix

Abra o **Zabbix** (botão no topo, `Admin` / `zabbix`):

- *Monitoramento → Incidentes*: **Agente Zabbix sem comunicação em loja-web-01**, severidade Alta.
- *Monitoramento → Hosts*: o ícone **ZBX** do loja-web-01 está vermelho. Passe o mouse sobre ele para ver o erro.

## 2. Teste como o Zabbix testa

```bash
zabbix_get -s 127.0.0.1 -k agent.ping
systemctl status zabbix-agent2 --no-pager
```

```
zabbix_get: Get value error: cannot read from socket: Connection reset by peer
● zabbix-agent2.service ... active (running)
```

O agente está rodando, mas derruba a conexão.

## 3. Leia o log do agente

```bash
sudo tail -n 5 /var/log/zabbix/zabbix_agent2.log
```

```
connection from "127.0.0.1" rejected, allowed hosts: "10.20.0.50"
```

O agente só aceita o endereço 10.20.0.50.

## 4. Corrija a configuração

```bash
sudo grep -nE '^(Server|ServerActive|Hostname)=' /etc/zabbix/zabbix_agent2.conf
head -1 /etc/zabbix/zabbix_agent2.conf
```

O arquivo foi aplicado pelo perfil "filial" do Ansible, que aponta para o proxy zbx-proxy-01. Aqui, o Zabbix Server é `127.0.0.1`:

```bash
sudo sed -i -e 's/^Server=.*/Server=127.0.0.1/' \
            -e 's/^ServerActive=.*/ServerActive=127.0.0.1/' /etc/zabbix/zabbix_agent2.conf
sudo systemctl restart zabbix-agent2
zabbix_get -s 127.0.0.1 -k agent.ping      # 1
```

> Não use `Server=0.0.0.0/0`: isso deixa o agente responder para qualquer um na rede.

Em até 30 segundos, o ícone ZBX fica verde e o incidente do agente aparece como **Resolvido**.

## 5. O que estava escondido

Volte em *Monitoramento → Incidentes*:

```
Pouco espaço em /var/log em loja-web-01 (94 %)
```

Esse disco encheu **enquanto o monitoramento estava cego**. Ninguém foi avisado porque não havia dados para o trigger avaliar.

```bash
df -h /var/log
sudo du -sh /var/log/* | sort -h | tail -3
ls -lh /var/log/auditoria/
cat /etc/cron.d/auditoria-export
```

Um export de auditoria de 190 MB, gravado em `/var/log` por um job de compliance.

## 6. Libere o espaço

Combine com o dono do export (compliance). No lab, apague ou mova o arquivo:

```bash
sudo rm /var/log/auditoria/export-completo-2026-10.json
df -h /var/log
```

Em seguida, o incidente de disco também fica **Resolvido** no Zabbix. Como prevenção, o export deveria ir para uma partição própria (ou para um bucket), não para `/var/log`.
