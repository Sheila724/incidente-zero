#!/usr/bin/env bash
# Cenário 007 — valida a correção. Exit 0 = resolvido.

status=0
pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; status=1; }
contorno() { echo "⚠️ $1"; status=1; }   # saída que só esconde o sintoma

SSHOPTS="-o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR -o ConnectTimeout=5"

systemctl is-active --quiet ssh && pass "Serviço ssh ativo" || fail "O serviço ssh não está ativo"

# Sem atalhos que enfraquecem o SSH
cfg=$(sshd -T 2>/dev/null)
if echo "$cfg" | grep -qx 'strictmodes no'; then
  contorno "StrictModes foi desligado: isso remove a proteção para todos os usuários"
else
  pass "StrictModes continua ligado"
fi
if echo "$cfg" | grep -qx 'passwordauthentication yes'; then
  contorno "Login por senha foi habilitado: o pipeline deve entrar por chave"
else
  pass "Login por senha continua desligado"
fi

# Permissões do caminho
if [ -z "$(find /home/deploy /home/deploy/.ssh -maxdepth 0 -perm /022 2>/dev/null)" ]; then
  pass "Home e ~/.ssh do deploy sem escrita para grupo e outros"
else
  fail "A home ou a ~/.ssh do deploy ainda pode ser alterada por outros usuários"
fi
ak=/home/deploy/.ssh/authorized_keys
if [ "$(stat -c %U "$ak" 2>/dev/null)" = "deploy" ] && [ -z "$(find "$ak" -perm /022 2>/dev/null)" ]; then
  pass "authorized_keys pertence ao deploy e não é gravável por outros"
else
  fail "authorized_keys com dono ou permissões incorretos ($(stat -c '%U %a' "$ak" 2>/dev/null))"
fi

# O teste que importa: o pipeline consegue entrar e executar o deploy
if out=$(ssh $SSHOPTS -i /home/aluno/ci/deploy_key deploy@127.0.0.1 'bash app/deploy.sh 2026.10.17-1' 2>&1); then
  pass "Pipeline entra como deploy e executa o deploy.sh"
else
  fail "O login do pipeline ainda falha: $(echo "$out" | tail -1)"
fi

# Bônus: nada mais gravável por todos no home do deploy
if [ -z "$(find /home/deploy -perm -o+w 2>/dev/null | head -1)" ]; then
  pass "Bônus: nenhum arquivo do deploy gravável por qualquer usuário"
fi

exit $status
