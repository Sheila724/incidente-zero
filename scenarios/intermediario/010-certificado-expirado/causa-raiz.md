# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| No servidor antigo | Timer de renovação ativo; um cron noturno de "manutenção" fazia `reload` do nginx todo dia |
| Mês passado | Migração para o template novo: scripts e units copiados, timer não habilitado, cron de manutenção não migrado |
| 91 dias atrás | Último certificado emitido (validade de 90 dias) |
| 30 dias atrás | Começa a janela de renovação. Nada acontece |
| Ontem, 18h | O certificado vence. Navegadores e o app passam a recusar a conexão |
| Hoje, 08:02 | O monitoramento alerta e o suporte acumula reclamações |

## Os 5 porquês

1. **Por que os clientes não conseguiam acessar?** O certificado servido estava vencido.
2. **Por que não foi renovado?** O timer de renovação estava desabilitado.
3. **Por que estava desabilitado?** A migração copiou os arquivos, mas não habilitou as units.
4. **Por que renovar na mão não resolveu na hora?** O script não recarrega o nginx; no servidor antigo, um cron de manutenção fazia isso "por acaso".
5. **Por que ninguém percebeu em 30 dias?** Não havia alerta de vencimento próximo, só de certificado já vencido.

## Fatores que contribuíram

- Uma dependência invisível (o reload do cron de manutenção) que ninguém sabia que existia.
- Migração validada por "o site abre", sem checar os processos agendados.
- Alerta só no vencimento, sem antecedência.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Alerta quando faltar 21 dias, e crítico com 7 dias | Detecção |
| Reload do nginx como parte da renovação (deploy hook) | Prevenção |
| Checklist de migração: `systemctl list-unit-files --state=enabled` antes e depois | Processo |
| Teste periódico da renovação em homologação (`--force`) | Prevenção |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| `openssl s_client` mostrando o `notAfter` de ontem no certificado servido | Evidência |
| O timer `renova-certificado` desabilitado e o script sem `reload` do nginx | Evidência da causa |
| O arquivo novo em `/etc/ssl/loja/` com datas válidas | Coincidência: o nginx continua entregando o certificado antigo até o `reload` |

## Para o seu postmortem

Certificado vencido é um dos incidentes mais evitáveis que existem, e mesmo assim derruba empresas grandes todo ano. A falha quase nunca é "esqueceram de renovar": é a automação que parou em silêncio. Monitore a **data de validade do que está sendo servido**, não só se o site responde.
