#!/usr/bin/env bash
# Cenário 007 — injeta a falha:
# rotação da chave copiada como root + chmod -R 777 no home do deploy.
set -euo pipefail

# Rotação da chave do pipeline: nova chave, copiada com sudo cp (dono vira root)
rm -f /home/aluno/ci/deploy_key /home/aluno/ci/deploy_key.pub
ssh-keygen -q -t ed25519 -N '' -C 'pipeline-deploy-2026Q4' -f /home/aluno/ci/deploy_key
chown aluno:aluno /home/aluno/ci/deploy_key /home/aluno/ci/deploy_key.pub
install -o root -g root -m 0600 /home/aluno/ci/deploy_key.pub /tmp/nova_chave.pub
cp /tmp/nova_chave.pub /home/deploy/.ssh/authorized_keys
chown root:root /home/deploy/.ssh/authorized_keys
rm -f /tmp/nova_chave.pub

# "Resolvendo um problema de permissão" para subir um arquivo
chmod -R 777 /home/deploy

# Autoverificação: o login por chave precisa falhar
if ssh -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
       -i /home/aluno/ci/deploy_key deploy@127.0.0.1 true 2>/dev/null; then
  echo "inject: o login ainda funciona, a falha não foi reproduzida" >&2
  exit 1
fi
