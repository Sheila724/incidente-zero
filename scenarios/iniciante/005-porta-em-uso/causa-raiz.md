# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| Semana passada | `debug-http` criado e habilitado para investigar o chamado #4512 |
| Desde então | O serviço falha a cada boot sem ninguém notar (o nginx sempre chegava antes) |
| 03:00 | Reboot para atualização; desta vez o `debug-http` ocupa a porta 80 primeiro |
| 03:00 | O nginx falha com "Address already in use" |
| 03:00–08:05 | A raiz do servidor fica exposta na internet |
| 08:02 | Cliente abre chamado; o monitoramento não alertou |

## Os 5 porquês

1. **Por que o site mostrava arquivos?** Um servidor de testes estava na porta 80 servindo `/`.
2. **Por que ele estava na porta 80?** Subiu antes do nginx no boot.
3. **Por que ele subia no boot?** Foi habilitado (`enable`) durante uma investigação e nunca removido.
4. **Por que ninguém percebeu antes?** Nos boots anteriores o nginx ganhava a disputa, e a falha do debug ficava só no log.
5. **Por que o monitoramento não alertou?** Ele só confere o código HTTP, e a listagem também respondia 200.

## Fatores que contribuíram

- Ferramenta de teste rodando como root, servindo a raiz do sistema.
- Nenhum inventário de serviços habilitados no servidor.
- Monitoramento sem verificação de conteúdo.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| Remover o `debug-http` e auditar serviços habilitados (`systemctl list-unit-files --state=enabled`) | Correção |
| Monitoramento verificar um texto esperado na página, não só o status 200 | Detecção |
| Servidores de teste nunca em produção, nunca como root, nunca habilitados | Processo |
| Tratar a exposição como incidente de segurança: revisar acessos e rotacionar segredos | Segurança |

## Para o seu postmortem

Este incidente tem duas faces: **disponibilidade** (a loja fora do ar) e **segurança** (dados expostos). Um bom postmortem trata as duas, e a segunda costuma exigir ações que vão além do servidor.
