# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 10:28 | Script de hardening (SEC-142) e timer publicados no servidor |
| 10:30 | Primeira execução: `chmod -R o-rwx /var/www/loja` |
| 10:30 | O nginx (`www-data`) perde acesso; home em 403, demais páginas em 404 |
| 10:32 | O monitoramento alerta HTTP 403 |
| a cada 2 min | O timer reaplica a mudança, desfazendo correções manuais |

## Os 5 porquês

1. **Por que a loja deu 403?** O nginx não conseguia ler os arquivos.
2. **Por que não conseguia?** O `www-data` caía em "outros", e "outros" perdeu todas as permissões.
3. **Por que "outros" perdeu as permissões?** O script de hardening removeu `o-rwx` de tudo.
4. **Por que o script não previu o nginx?** Ele foi escrito pensando em "usuários do servidor", sem considerar que o próprio serviço web é um deles.
5. **Por que isso chegou em produção?** O script não foi testado num ambiente com o site rodando, e não havia verificação do site após aplicá-lo.

## Fatores que contribuíram

- A falha voltava a cada 2 minutos, o que dá a falsa impressão de "resolvido" depois do primeiro `chmod`.
- O 404 nas páginas internas esconde a causa: parece que os arquivos sumiram.
- Mudança de segurança tratada como "não mexe em nada visível", sem plano de rollback.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Arquivos do site com grupo `www-data` e `u=rwX,g=rX,o=` | Correção |
| Testar mudanças de permissão em homologação com o serviço rodando | Prevenção |
| Smoke test (`curl`) ao fim de scripts que mexem em arquivos servidos | Detecção |
| Documentar tarefas agendadas no runbook do servidor | Processo |

## Para o seu postmortem

Segurança e disponibilidade não são opostos: o requisito da auditoria era legítimo e a solução final atende os dois. O aprendizado é **envolver quem opera o serviço** ao desenhar controles de segurança.
