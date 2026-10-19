# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Meses atrás | INC-1877: a `location /api/` ganha um timeout próprio, com o endereço da aplicação escrito direto |
| Semana passada | O time decide mover a aplicação para a 8081 (DEP-87) e abre a tarefa de ajustar o proxy (DEP-88) |
| 19:00 | O marketing dispara o e-mail da promoção |
| 19:01 | Deploy 2.4.0: a aplicação reinicia na 8081. O DEP-88 não foi feito |
| 19:01 | O nginx continua mandando tudo para a 8080: 502 em todas as rotas |
| 19:04 | O monitoramento alerta |

## Os 5 porquês

1. **Por que a loja dava 502?** O nginx encaminhava para a porta 8080, onde não havia mais nada.
2. **Por que nada na 8080?** O deploy 2.4.0 mudou a aplicação para a 8081.
3. **Por que o nginx não acompanhou?** A mudança do proxy era uma tarefa separada, que não entrou no deploy.
4. **Por que o pipeline ficou verde?** Ele conferia se o serviço `app` estava ativo, não se o site respondia pelo nginx.
5. **Por que corrigir o upstream não bastava?** O endereço estava duplicado: a `location /api/` tinha o próprio `proxy_pass`.

## Fatores que contribuíram

- Mudança de infraestrutura (porta) acoplada a um deploy de funcionalidade.
- Configuração do proxy fora do repositório da aplicação, sem revisão conjunta.
- Endereço do backend repetido em mais de um lugar.
- Deploy no mesmo minuto do disparo da campanha.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Smoke test pós-deploy pelo nginx, em todas as rotas críticas (`/`, `/api/`) | Detecção |
| Configuração do nginx versionada junto com a aplicação | Prevenção |
| Uma única definição do backend (`upstream`) | Prevenção |
| Mudanças de porta compatíveis: aplicação escuta nas duas até o proxy migrar | Prevenção |
| Não fazer deploy em cima do horário de campanha | Processo |

## Para o seu postmortem

"O serviço está ativo" não é o mesmo que "o usuário consegue usar". Toda verificação de deploy deveria passar pelo mesmo caminho que o cliente percorre.
