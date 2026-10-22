# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Meses atrás | INC-2291: DNS interno instável. Alguém fixa o IP da API no `/etc/hosts` da loja como contorno |
| Meses atrás | O DNS é estabilizado, mas a linha no `/etc/hosts` fica |
| 02:00 | Migração da API: pag-api-01 (10.20.0.9) é desligado, pag-api-02 (10.20.0.15) assume |
| 02:00 | O DNS interno é atualizado; os testes do time de pagamentos passam |
| 02:05 | Todo checkout da loja dá timeout para 10.20.0.9 e responde 502 |
| 02:14 | O monitoramento alerta |

## Os 5 porquês

1. **Por que o checkout falhou?** A aplicação chamava a API de pagamentos no IP antigo, já desligado.
2. **Por que usava o IP antigo, se o DNS estava certo?** O `/etc/hosts` fixava o IP antigo, e ele é consultado antes do DNS.
3. **Por que havia um IP fixado?** Um contorno temporário de um incidente antigo nunca foi desfeito.
4. **Por que ninguém percebeu antes da migração?** Os testes da migração foram feitos do lado da API, não do lado de quem a consome.
5. **Por que corrigir o `/etc/hosts` não bastou?** A aplicação resolve o nome uma vez, ao iniciar, e guarda o IP.

## Fatores que contribuíram

- Contornos "temporários" sem dono nem prazo para serem desfeitos.
- Migração sem checklist dos consumidores da API.
- Aplicação que ignora o TTL do DNS.
- Ninguém monitorava o checkout de ponta a ponta, só cada peça separada.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Todo contorno vira um ticket com dono e data para ser desfeito | Processo |
| Auditoria periódica de entradas no `/etc/hosts` dos servidores | Detecção |
| Aplicação respeitando o TTL do DNS (ou resolvendo a cada conexão) | Prevenção |
| Teste sintético do checkout completo no monitoramento | Detecção |
| Checklist de migração com os consumidores da API | Processo |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| Log da aplicação tentando `10.20.0.9`, `getent` respondendo `10.20.0.9` e `dig` respondendo `10.20.0.15` | Evidência |
| A linha esquecida no `/etc/hosts` | Evidência da causa |
| O checkout voltar depois do `restart` da aplicação | Parcial: só funciona porque o `/etc/hosts` já foi corrigido; sozinho, o restart resolve para o IP errado de novo |

## Para o seu postmortem

Os dois times tinham razão do seu ponto de vista: o DNS estava certo, e a loja não tinha mudado nada. O incidente mora **entre** os dois. Por isso vale testar dependências do ponto de vista de quem consome, não só de quem fornece.
