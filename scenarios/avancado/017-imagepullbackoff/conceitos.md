# Conceitos para este cenário

## Como um Deployment atualiza a aplicação

Cada versão do template do Deployment ganha um **ReplicaSet**. No *rolling update* (o padrão), o Kubernetes sobe pods do ReplicaSet novo e só remove os do antigo quando os novos ficam prontos:

```
ReplicaSet antigo (2.5):  ● ●      → ● ●  (continua servindo)
ReplicaSet novo   (2.6):           → ✖    (nunca fica pronto)
```

Os limites vêm de `maxSurge` (quantos pods a mais podem existir) e `maxUnavailable` (quantos podem ficar indisponíveis). Com 2 réplicas e os valores padrão (25%), o Kubernetes cria 1 pod novo por vez e não derruba nenhum antigo antes disso.

Se o rollout não avança dentro de `progressDeadlineSeconds`, o Deployment marca `ProgressDeadlineExceeded`. Ele **não desfaz** nada sozinho: fica parado no meio.

```bash
kubectl rollout status deploy/loja-web
kubectl get rs                       # um ReplicaSet por versão
kubectl rollout history deploy/loja-web
```

## Imagens, registry e tags

Uma imagem é identificada por `registry/repositório:tag`:

```
registry.loja.local/loja-web:2.6
└── registry ──────┘└ repo ─┘└tag┘
```

| Estado | Significado |
|---|---|
| `ErrImagePull` | A tentativa de baixar a imagem falhou agora |
| `ImagePullBackOff` | Falhou antes; o kubelet espera cada vez mais para tentar de novo |

Os eventos do pod dizem o motivo: `not found` (nome ou tag não existe), `unauthorized` (falta credencial), `no such host` / `i/o timeout` (registry inacessível).

## Perguntando à registry

Registries seguem uma API HTTP padrão:

```bash
curl -s http://registry.loja.local:5000/v2/_catalog            # repositórios
curl -s http://registry.loja.local:5000/v2/loja-web/tags/list  # tags de um repositório
```

## Corrigir ou desfazer?

`kubectl rollout undo` volta o template para a versão anterior. É uma boa mitigação quando a versão nova está quebrando o serviço. Aqui o serviço está de pé e o objetivo é **entregar** a versão nova, então a correção é apontar para a imagem certa.
