# Incidente Zero — imagem base dos labs
# Ubuntu 24.04 com systemd, para que os cenários tenham systemctl/journalctl reais.

FROM ubuntu:24.04

LABEL org.opencontainers.image.source="https://github.com/Sheila724/incidente-zero" \
      org.opencontainers.image.description="Imagem base dos labs do Incidente Zero" \
      org.opencontainers.image.licenses="MIT"

ENV DEBIAN_FRONTEND=noninteractive \
    container=docker \
    LANG=C.UTF-8

RUN apt-get update && apt-get install -y --no-install-recommends \
      systemd systemd-sysv dbus \
      sudo ca-certificates curl wget \
      nginx python3 \
      vim nano less \
      procps psmisc htop lsof ncdu tree \
      iproute2 iputils-ping dnsutils netcat-openbsd \
      logrotate cron openssh-server openssh-client \
      mariadb-server mariadb-client python3-pymysql \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /var/lib/mysql/*

# MariaDB fica desligado e sem dados: só os cenários de banco o habilitam,
# criando o próprio diretório de dados (mariadb-install-db)
RUN systemctl disable mariadb 2>/dev/null || true

# Units que não fazem sentido dentro de container
RUN systemctl mask \
      systemd-udevd.service systemd-udevd-kernel.socket systemd-udevd-control.socket \
      getty.target console-getty.service systemd-firstboot.service \
      systemd-remount-fs.service 2>/dev/null || true

# Journal só em memória: os cenários controlam o que vai para /var/log
RUN mkdir -p /etc/systemd/journald.conf.d \
 && printf '[Journal]\nStorage=volatile\nRuntimeMaxUse=32M\n' \
      > /etc/systemd/journald.conf.d/lab.conf

# /var/log pode ser um tmpfs vazio: recria o diretório do nginx no boot
RUN printf 'd /var/log/nginx 0755 www-data adm -\n' > /etc/tmpfiles.d/lab-nginx.conf

# Em containers, o kernel pode recusar a gravação do loginuid, e o cron
# deixaria de executar os jobs ("Cannot make/remove an entry for the
# specified session"). O módulo fica opcional para o cron.
RUN sed -i 's/^session\s\+required\s\+pam_loginuid.so/session    optional     pam_loginuid.so/' /etc/pam.d/cron

# Usuário do aluno, com sudo
RUN useradd -m -s /bin/bash aluno \
 && echo 'aluno ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/aluno \
 && chmod 440 /etc/sudoers.d/aluno

STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]
