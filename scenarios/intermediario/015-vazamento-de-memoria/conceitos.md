# Conceitos para este cenário

## Lendo um gráfico de memória

| Formato do gráfico | O que costuma ser |
|---|---|
| Sobe e desce com o movimento | Uso normal: a memória acompanha a carga |
| Degrau único, depois estável | Mudança de configuração ou de volume (ex.: um cache que encheu) |
| Rampa que só sobe | **Vazamento:** algo é guardado e nunca liberado |
| Rampa, queda brusca, rampa de novo | Vazamento com reinícios (o processo morre e volta) |

O gráfico mostra **quando** começou. Cruzar esse horário com deploys e mudanças é, muitas vezes, metade da investigação.

## Memória de um processo no Linux

```bash
ps -o pid,rss,vsz,cmd -u app
systemctl status app            # mostra "Memory:" do serviço
```

| Medida | Significado |
|---|---|
| **RSS** | Memória física realmente em uso pelo processo |
| **VSZ** | Memória virtual reservada (inclui o que nunca foi usado; costuma enganar) |

No Zabbix, `proc.mem[<processo>,<usuário>,,<linha de comando>,rss]` coleta o RSS de um processo.

## Limites no systemd

```ini
[Service]
MemoryMax=900M
```

Acima do limite, o kernel encerra o processo (OOM). Com `Restart=always`, ele volta, a memória zera e o ciclo recomeça: a loja cai por alguns segundos a cada volta. O limite protege o servidor, mas não corrige nada.

## Cache sem limite é vazamento

Um cache precisa de uma regra de saída: tamanho máximo, tempo de vida (TTL), ou os dois. Sem isso, ele cresce com o tráfego até ocupar toda a memória.

## Mitigar × corrigir

| Ação | Tipo | Efeito |
|---|---|---|
| Reiniciar o processo | Mitigação | Memória volta ao normal, mas o problema volta junto |
| Rollback do deploy | Mitigação | Perde a funcionalidade nova |
| Configurar o limite do cache | **Correção** | Mantém a funcionalidade e estabiliza a memória |

## Registrar no Zabbix

No Zabbix, cada incidente tem o botão **Atualizar**, para reconhecer o incidente (*acknowledge*), escrever uma mensagem e mudar a severidade. O histórico fica junto com o evento: quem atender o próximo alerta parecido já sabe o que foi feito.
