# Causa raiz e prevenção

## Linha do tempo

| Quando | Evento |
|---|---|
| Ontem, 15:10 | Rotação da chave do pipeline: `sudo cp` da nova chave para o `authorized_keys` (dono vira `root`) |
| Ontem, 16:40 | Suporte roda `chmod -R 777 /home/deploy` para conseguir subir um arquivo |
| Ontem, 17:00 | Primeiro deploy falha com `Permission denied (publickey)` |
| Hoje, 11:20 | Correção urgente travada; incidente aberto |

## Os 5 porquês

1. **Por que o deploy falhou?** O sshd recusou a chave do pipeline.
2. **Por que recusou?** As permissões e o dono do caminho até o `authorized_keys` não eram seguros.
3. **Por que estavam assim?** Um `chmod -R 777` no home e uma cópia de chave feita como root.
4. **Por que alguém usou 777?** Para contornar rápido um problema de permissão, sem entender a causa.
5. **Por que demorou para identificar?** Duas mudanças no mesmo dia e o erro do cliente sem detalhes.

## Fatores que contribuíram

- Rotação de credencial sem um passo de validação (testar o login logo depois).
- Mudanças manuais em produção, sem registro.
- O primeiro palpite foi "a chave nova está errada", o que desviou a investigação.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Rotação de chaves por automação (Ansible) com teste de login no final | Prevenção |
| Proibir `chmod 777` em produção (revisão e auditoria) | Processo |
| Alerta de arquivos graváveis por todos em `/home` e `/etc` | Detecção |
| Runbook do erro `Permission denied (publickey)`: começar pelo log do servidor | Resposta |

## Evidência × coincidência

Num postmortem, separe o que **confirmou** a causa do que só aconteceu junto com a recuperação. É isso que impede que um reinício seja tratado como diagnóstico.

| O que foi observado | Peso |
|---|---|
| Log do `sshd` com `Authentication refused: bad ownership or modes` | Evidência |
| `namei -l` mostrando a home ou a `~/.ssh` gravável por outros | Evidência |
| O login funcionar com `StrictModes no` | Coincidência perigosa: entra, mas desliga a proteção para todos |

## Para o seu postmortem

Quando duas mudanças acontecem perto do incidente, resista a culpar a mais óbvia. Aqui a rotação da chave era inocente quanto ao conteúdo, mas culpada quanto ao **dono** do arquivo. O log do servidor separa as hipóteses em segundos.
