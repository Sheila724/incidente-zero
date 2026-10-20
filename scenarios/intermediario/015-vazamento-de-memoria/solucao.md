# Solução comentada

> Tente resolver antes de ler. Se travar, use as dicas primeiro.

## 1. Veja o incidente no Zabbix

Abra o **Zabbix** (`Admin` / `zabbix`), em *Monitoramento → Incidentes*:

```
Média   Loja Exemplo usando muita memória em loja-web-01 (330 MB)
```

## 2. Olhe a tendência

Em *Monitoramento → Dados recentes*, filtre o host `loja-web-01` e clique em **Gráfico** no item *Loja Exemplo: memória do processo*.

A linha sobe sem parar, desde o horário do deploy, e nunca desce. É o desenho de um vazamento.

## 3. Confirme no servidor

```bash
systemctl status app --no-pager | grep -i memory
ps -o pid,rss,cmd -u app
sleep 30; ps -o pid,rss,cmd -u app
```

O RSS cresce alguns megabytes por minuto. O limite do serviço é `MemoryMax=900M`.

## 4. A armadilha: só reiniciar

```bash
sudo systemctl restart app
```

A memória cai para ~30 MB, o incidente resolve... e a rampa recomeça. Em alguns minutos o alerta volta.

## 5. Descubra o que mudou

```bash
cat /opt/app/CHANGELOG.md
cat /etc/app/app.env
curl -s http://127.0.0.1:8080/health
```

```
SESSION_CACHE_MAX=0      ← 0 = sem limite
"sessoes_em_cache": 7342 ← cresce a cada visita
```

A 2.5.0 guarda cada sessão em memória (~50 KB), e a configuração saiu sem limite.

## 6. Corrija sem desfazer o deploy

```bash
sudo sed -i 's/^SESSION_CACHE_MAX=0/SESSION_CACHE_MAX=2000/' /etc/app/app.env
sudo systemctl restart app
curl -s http://127.0.0.1:8080/health
```

Agora o cache guarda no máximo 2000 sessões (~100 MB) e descarta as mais antigas. No gráfico do Zabbix, a memória sobe até esse patamar e **estabiliza**.

## 7. Registre no Zabbix

Em *Monitoramento → Incidentes*, no incidente de memória (agora *Resolvido*), clique em **Atualizar**:

- Mensagem: `Cache de sessões sem limite (SESSION_CACHE_MAX=0) na 2.5.0. Ajustado para 2000 e loja reiniciada.`
- Marque **Reconhecer** e confirme.

Se o incidente já saiu da lista, mude o filtro para **Mostrar: Histórico**.
