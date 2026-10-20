# Conceitos para este cenário

## Como o Zabbix coleta dados de um servidor

O **agente Zabbix** roda no servidor monitorado e responde perguntas como "quanto de disco está em uso?". Existem dois modos:

| Modo | Quem inicia a conexão | Parâmetro no agente |
|---|---|---|
| **Passivo** | O Zabbix Server (ou proxy) pergunta ao agente, na porta 10050 | `Server=` |
| **Ativo** | O agente busca a lista de itens e envia os dados para o servidor, na porta 10051 | `ServerActive=` |

O `Server=` funciona como uma **lista de permissão**: o agente só responde a quem está nela. Qualquer outro endereço é recusado.

## "Sem dados" é diferente de "tudo bem"

Quando o agente para de responder:

- os itens param de receber valores;
- triggers que dependem desses valores **não disparam**, porque não há dados para avaliar;
- só triggers como `nodata()` percebem a ausência.

É a **falha silenciosa**: o painel não mostra disco cheio, CPU alta ou serviço parado, não porque está tudo bem, mas porque ninguém está olhando.

## Ferramentas

| Ferramenta | Para quê |
|---|---|
| `zabbix_get -s <ip> -k <chave>` | Faz a mesma pergunta que o servidor faz ao agente (modo passivo) |
| `zabbix_agent2 -t <chave>` | Testa uma chave localmente, sem passar pela rede |
| `/var/log/zabbix/zabbix_agent2.log` | Conexões recusadas, erros de configuração |
| `zbx-api` (neste lab) | Consulta a API do Zabbix pelo terminal, ex.: `zbx-api problem.get '{}'` |

## No frontend do Zabbix

| Onde | O que ver |
|---|---|
| Monitoramento → Incidentes | Problemas abertos, severidade e duração |
| Monitoramento → Hosts | Disponibilidade (ícone **ZBX** verde, vermelho ou cinza) |
| Monitoramento → Dados recentes | Último valor de cada item e o gráfico |

## Gerência de configuração

Ferramentas como o Ansible aplicam a mesma configuração em muitos servidores de uma vez. Um perfil errado (por exemplo, o da filial aplicado na matriz) quebra **todos** de uma vez, e "tudo verde" no Ansible só quer dizer que o arquivo foi copiado, não que o serviço funciona.
