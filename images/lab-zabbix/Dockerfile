# Incidente Zero — imagem dos labs com Zabbix 7.0 LTS
# Zabbix Server + frontend (nginx + php-fpm) + agente 2, com o banco já criado.
# Tudo fica desligado: os cenários de Zabbix ligam o que precisam.

ARG BASE=ghcr.io/sheila724/incidente-zero/lab-base:latest
FROM ${BASE}

LABEL org.opencontainers.image.source="https://github.com/Sheila724/incidente-zero" \
      org.opencontainers.image.description="Imagem dos labs do Incidente Zero com Zabbix 7.0 LTS" \
      org.opencontainers.image.licenses="MIT"

ARG TARGETARCH
ARG ZABBIX_VERSION=7.0

# Repositório oficial do Zabbix (amd64 e arm64)
RUN set -eux; \
    case "${TARGETARCH:-amd64}" in \
      arm64) repo="ubuntu-arm64" ;; \
      *)     repo="ubuntu" ;; \
    esac; \
    curl -fsSL -o /tmp/zabbix-release.deb \
      "https://repo.zabbix.com/zabbix/${ZABBIX_VERSION}/${repo}/pool/main/z/zabbix-release/zabbix-release_latest_${ZABBIX_VERSION}+ubuntu24.04_all.deb"; \
    dpkg -i /tmp/zabbix-release.deb; rm /tmp/zabbix-release.deb; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
      zabbix-server-mysql zabbix-frontend-php zabbix-nginx-conf zabbix-sql-scripts \
      zabbix-agent2 zabbix-get zabbix-sender locales jq; \
    sed -i 's/^# *pt_BR.UTF-8/pt_BR.UTF-8/' /etc/locale.gen; locale-gen >/dev/null; \
    apt-get clean; rm -rf /var/lib/apt/lists/*

# MariaDB enxuto para o lab
RUN printf '%s\n' '[mysqld]' \
      'bind-address            = 127.0.0.1' \
      'innodb_buffer_pool_size = 64M' \
      'innodb_log_file_size    = 16M' \
      'performance_schema      = OFF' \
      'character-set-server    = utf8mb4' \
      'collation-server        = utf8mb4_bin' \
      'innodb_use_native_aio   = 0' \
      > /etc/mysql/mariadb.conf.d/60-zabbix.cnf

# Banco do Zabbix criado na imagem: o lab sobe em segundos
RUN set -eux; \
    install -d -o mysql -g mysql -m 0750 /var/lib/mysql; \
    install -d -o mysql -g mysql -m 0755 /run/mysqld; \
    mariadb-install-db --user=mysql --datadir=/var/lib/mysql --skip-test-db >/dev/null; \
    (mariadbd --user=mysql --datadir=/var/lib/mysql --skip-networking --socket=/run/mysqld/mysqld.sock &); \
    for _ in $(seq 1 60); do mariadb -e 'SELECT 1' >/dev/null 2>&1 && break; sleep 1; done; \
    mariadb -e "CREATE DATABASE zabbix CHARACTER SET utf8mb4 COLLATE utf8mb4_bin; \
                CREATE USER zabbix@localhost IDENTIFIED BY 'zabbix'; \
                GRANT ALL PRIVILEGES ON zabbix.* TO zabbix@localhost; \
                SET GLOBAL log_bin_trust_function_creators = 1;"; \
    zcat /usr/share/zabbix-sql-scripts/mysql/server.sql.gz | mariadb --default-character-set=utf8mb4 zabbix; \
    mariadb zabbix -e "UPDATE users SET lang='pt_BR', timezone='America/Sao_Paulo' WHERE username='Admin';"; \
    mariadb -e "SET GLOBAL log_bin_trust_function_creators = 0;"; \
    mariadb-admin shutdown; \
    sleep 2

# Zabbix Server: poucos processos, intervalos curtos (incidentes aparecem rápido)
RUN set -eux; \
    f=/etc/zabbix/zabbix_server.conf; \
    sed -i 's/^# DBPassword=.*/DBPassword=zabbix/' "$f"; \
    printf '%s\n' \
      'StartPollers=3' 'StartPollersUnreachable=1' 'StartPingers=1' 'StartTrappers=2' \
      'StartDiscoverers=1' 'StartHTTPPollers=1' 'StartPreprocessors=2' \
      'StartAgentPollers=1' 'StartHTTPAgentPollers=1' 'StartSNMPPollers=0' \
      'CacheSize=16M' 'HistoryCacheSize=8M' 'ValueCacheSize=16M' 'TrendCacheSize=4M' \
      'CacheUpdateFrequency=5' 'UnreachablePeriod=20' 'UnreachableDelay=5' 'UnavailableDelay=20' \
      >> "$f"

# Frontend: nginx na porta 8081, sem assistente de instalação
RUN set -eux; \
    sed -i -e 's/^#\?\s*listen\s.*/        listen          8081;/' \
           -e 's/^#\?\s*server_name\s.*/        server_name     _;/' /etc/zabbix/nginx.conf; \
    sed -i 's|^;\?\s*php_value\[date.timezone\].*|php_value[date.timezone] = America/Sao_Paulo|' \
      /etc/zabbix/php-fpm.conf; \
    { echo '<?php'; \
      echo "\$DB['TYPE'] = 'MYSQL'; \$DB['SERVER'] = 'localhost'; \$DB['PORT'] = '0';"; \
      echo "\$DB['DATABASE'] = 'zabbix'; \$DB['USER'] = 'zabbix'; \$DB['PASSWORD'] = 'zabbix';"; \
      echo "\$DB['SCHEMA'] = ''; \$DB['ENCRYPTION'] = false; \$DB['KEY_FILE'] = ''; \$DB['CERT_FILE'] = '';"; \
      echo "\$DB['CA_FILE'] = ''; \$DB['VERIFY_HOST'] = false; \$DB['CIPHER_LIST'] = ''; \$DB['VAULT'] = '';"; \
      echo "\$DB['VAULT_URL'] = ''; \$DB['VAULT_PREFIX'] = ''; \$DB['VAULT_DB_PATH'] = ''; \$DB['VAULT_TOKEN'] = '';"; \
      echo "\$DB['VAULT_CERT_FILE'] = ''; \$DB['VAULT_KEY_FILE'] = ''; \$DB['DOUBLE_IEEE754'] = true;"; \
      echo "\$ZBX_SERVER_NAME = 'Incidente Zero';"; \
      echo "\$IMAGE_FORMAT_DEFAULT = IMAGE_FORMAT_PNG;"; \
    } > /etc/zabbix/web/zabbix.conf.php; \
    chown www-data:www-data /etc/zabbix/web/zabbix.conf.php; chmod 0640 /etc/zabbix/web/zabbix.conf.php

# Ferramentas do lab: API pela linha de comando e subida dos serviços
COPY zbx-api zabbix-lab-up zabbix-lab-host /usr/local/bin/

# Tudo desligado por padrão
RUN systemctl disable zabbix-server zabbix-agent2 php8.3-fpm mariadb 2>/dev/null; rm -f /etc/nginx/sites-enabled/default
