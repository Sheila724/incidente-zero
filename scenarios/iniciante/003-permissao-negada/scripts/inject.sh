#!/usr/bin/env bash
# Cenário 003 — injeta a falha:
# um "hardening" de segurança agendado tira o acesso do nginx aos arquivos do site.
set -euo pipefail

cat > /usr/local/bin/hardening-web.sh <<'EOF'
#!/usr/bin/env bash
# Hardening do diretório web — pedido da auditoria de segurança (ticket SEC-142).
# Objetivo: arquivos do site não podem ser lidos por "qualquer usuário" do servidor.
chown -R deploy:deploy /var/www/loja
chmod -R o-rwx /var/www/loja
EOF
chmod 0755 /usr/local/bin/hardening-web.sh

cat > /etc/systemd/system/hardening-web.service <<'EOF'
[Unit]
Description=Hardening de permissões do diretório web (SEC-142)

[Service]
Type=oneshot
ExecStart=/usr/local/bin/hardening-web.sh
EOF

cat > /etc/systemd/system/hardening-web.timer <<'EOF'
[Unit]
Description=Executa o hardening do diretório web a cada 2 minutos

[Timer]
OnActiveSec=2min
OnUnitActiveSec=2min

[Install]
WantedBy=timers.target
EOF

systemctl daemon-reload
systemctl start hardening-web.service
systemctl enable --now hardening-web.timer

# Autoverificação: a home precisa responder 403
code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/ || true)
if [ "$code" != "403" ]; then
  echo "inject: falha não reproduzida (HTTP $code)" >&2
  exit 1
fi
