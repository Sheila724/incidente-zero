#!/usr/bin/env bash
# Cenário 007 — prepara o ambiente SAUDÁVEL:
# usuário deploy com login só por chave; a chave do pipeline em ~aluno/ci.
set -euo pipefail

id deploy &>/dev/null || useradd -m -s /bin/bash deploy
install -d -o deploy -g deploy -m 0755 /home/deploy/app
echo "release 2026.10.14-3 (a7c91e2)" > /home/deploy/app/release.txt
cat > /home/deploy/app/deploy.sh <<'EOF'
#!/usr/bin/env bash
# Chamado pelo pipeline: atualiza a release da aplicação
set -euo pipefail
echo "release $1" > "$HOME/app/release.txt"
echo "deploy concluído: $1"
EOF
chown deploy:deploy /home/deploy/app/release.txt /home/deploy/app/deploy.sh
chmod 0644 /home/deploy/app/release.txt
chmod 0755 /home/deploy/app/deploy.sh
chmod 0755 /home/deploy

# SSH: só chave, sem senha
mkdir -p /etc/ssh/sshd_config.d
cat > /etc/ssh/sshd_config.d/10-lab.conf <<'EOF'
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin no
EOF
ssh-keygen -A >/dev/null

# Chave antiga do pipeline (antes da rotação)
install -d -o aluno -g aluno -m 0700 /home/aluno/ci
rm -f /home/aluno/ci/deploy_key /home/aluno/ci/deploy_key.pub
ssh-keygen -q -t ed25519 -N '' -C 'pipeline-deploy-antiga' -f /home/aluno/ci/deploy_key
install -d -o deploy -g deploy -m 0700 /home/deploy/.ssh
install -o deploy -g deploy -m 0600 /home/aluno/ci/deploy_key.pub /home/deploy/.ssh/authorized_keys
chown aluno:aluno /home/aluno/ci/deploy_key /home/aluno/ci/deploy_key.pub

systemctl enable ssh >/dev/null 2>&1 || true
systemctl restart ssh

for _ in $(seq 1 20); do
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
      -i /home/aluno/ci/deploy_key deploy@127.0.0.1 true 2>/dev/null && break
  sleep 0.5
done

# nginx com uma página simples, para o monitoramento HTTP
install -d -m 0755 /var/www/app
echo '<h1>Loja Exemplo</h1><p>Servidor de aplicação.</p>' > /var/www/app/index.html
cat > /etc/nginx/sites-available/app <<'EOF'
server {
    listen 80 default_server;
    root /var/www/app;
    index index.html;
}
EOF
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/app /etc/nginx/sites-enabled/app
nginx -t
systemctl enable nginx
systemctl restart nginx

ssh -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
    -i /home/aluno/ci/deploy_key deploy@127.0.0.1 true
