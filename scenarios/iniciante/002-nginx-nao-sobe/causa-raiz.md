# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 23:50 | Arquivo da landing page criado em `sites-available` e ativado |
| 23:55 | Deploy roda `systemctl restart nginx` |
| 23:55 | O nginx para, falha no teste de configuração e não volta |
| 23:55 | O script de deploy não confere o resultado e termina "com sucesso" |
| 00:07 | O monitoramento alerta: conexão recusada |

## Os 5 porquês

1. **Por que a loja caiu?** O nginx não estava rodando.
2. **Por que não estava rodando?** Falhou ao reiniciar por erro de configuração.
3. **Por que havia erro?** Um `;` faltando e um segundo `default_server` na mesma porta.
4. **Por que o erro chegou em produção?** A configuração não foi testada com `nginx -t` antes do restart.
5. **Por que o deploy disse "sucesso"?** O script não verificava se o serviço voltou nem se o site respondia.

## Fatores que contribuíram

- Uso de `restart` em vez de `reload`: com `reload`, a configuração antiga teria continuado no ar.
- Uma mudança só da promoção conseguiu derrubar **todos** os sites do servidor.
- Deploy em cima do horário de pico, sem margem para reagir.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Rodar `nginx -t` no pipeline antes de aplicar | Prevenção |
| Usar `nginx -t && systemctl reload nginx` | Prevenção |
| Teste pós-deploy (smoke test) com `curl` em cada site | Detecção |
| Congelar deploys perto de eventos críticos | Processo |

## Para o seu postmortem

O erro de digitação é o gatilho, não a causa. Qualquer pessoa esquece um `;`. A pergunta útil é: **por que o processo deixou um erro de digitação derrubar a loja inteira?**
