# Causa raiz e prevenção

## Linha do tempo

| Horário | Evento |
|---|---|
| 09:12 | O pipeline constrói e publica `registry.loja.local/loja-web:2.6` |
| 09:13 | O deploy aplica a imagem `loja-web:v2.6`, usando a tag do Git |
| 09:13 | O pod novo fica em `ErrImagePull` e depois em `ImagePullBackOff` |
| 09:13 em diante | Os 2 pods da 2.5 continuam atendendo (rolling update) |
| 09:15 | `ProgressDeadlineExceeded`: o pipeline falha e abre o alerta |

## Os 5 porquês

1. **Por que o deploy não terminou?** O pod da 2.6 nunca ficou pronto.
2. **Por que não ficou pronto?** O nó não conseguiu baixar a imagem.
3. **Por que não conseguiu?** A tag `v2.6` não existe na registry; a publicada é `2.6`.
4. **Por que tags diferentes?** O passo de build usa a versão sem `v`; o de deploy usa a tag do Git, com `v`.
5. **Por que ninguém notou antes?** Os dois passos montam o nome da imagem separadamente, e o pipeline não confere se a imagem existe antes de aplicar.

## Fatores que contribuíram

- O nome da imagem montado em dois lugares diferentes.
- Nenhuma verificação da imagem antes do deploy.
- O rolling update escondeu o problema dos clientes: bom para a loja, ruim para a percepção de urgência.

## Ações de prevenção

| Ação | Tipo |
|---|---|
| O deploy usa exatamente a referência gerada pelo build (de preferência o digest `@sha256:...`) | Prevenção |
| Passo no pipeline que confirma que a imagem existe antes do `kubectl apply` | Prevenção |
| Uma convenção única de tags, documentada | Prevenção |
| Alerta quando um Deployment fica com réplicas atualizadas abaixo do desejado | Detecção |

## Evidência × coincidência

| O que foi observado | Peso |
|---|---|
| Evento do pod: `loja-web:v2.6: not found` | Evidência |
| A listagem de tags da registry sem `v2.6` e o log do pipeline com os dois nomes | Evidência da causa |
| A loja continuar respondendo 200 | Consequência do rolling update, não sinal de que está tudo bem |
| O alerta sumir depois de um `rollout undo` | Coincidência: o deploy "termina", mas a 2.6 não vai ao ar |

## Para o seu postmortem

Um deploy travado não é um deploy que falhou de forma limpa: o cluster fica no meio do caminho, com duas versões convivendo. Saber ler `kubectl get rs` e os eventos do pod é o que permite decidir entre concluir e desfazer.
