# Incidente Zero — imagem dos labs com Kubernetes (k3s)
# k3s de nó único, com as imagens do cluster embutidas (funciona sem Docker Hub).
# O k3s fica desligado: os cenários sobem o cluster com `k3s-lab-up`.

ARG BASE=ghcr.io/sheila724/incidente-zero/lab-base:latest
FROM ${BASE}

LABEL org.opencontainers.image.source="https://github.com/Sheila724/incidente-zero" \
      org.opencontainers.image.description="Imagem dos labs do Incidente Zero com Kubernetes (k3s)" \
      org.opencontainers.image.licenses="MIT"

ARG TARGETARCH
ARG K3S_VERSION=v1.33.5+k3s1

RUN set -eux; \
    case "${TARGETARCH:-amd64}" in \
      arm64) bin="k3s-arm64"; imgs="k3s-airgap-images-arm64.tar.zst" ;; \
      *)     bin="k3s";       imgs="k3s-airgap-images-amd64.tar.zst" ;; \
    esac; \
    base="https://github.com/k3s-io/k3s/releases/download/$(echo "$K3S_VERSION" | sed 's/+/%2B/')"; \
    curl -fsSL -o /usr/local/bin/k3s "$base/$bin"; \
    chmod 0755 /usr/local/bin/k3s; \
    for c in kubectl crictl ctr; do ln -sf k3s "/usr/local/bin/$c"; done; \
    install -d /var/lib/rancher/k3s/agent/images; \
    curl -fsSL -o "/var/lib/rancher/k3s/agent/images/$imgs" "$base/$imgs"

# Cluster enxuto: sem Traefik e sem o balanceador do k3s; containerd sem overlay
# (o lab já roda dentro de um container)
RUN install -d /etc/rancher/k3s && printf '%s\n' \
      'write-kubeconfig-mode: "0644"' \
      'node-name: loja-k8s-01' \
      'disable:' '  - traefik' '  - servicelb' \
      'snapshotter: native' \
      'kubelet-arg:' '  - "eviction-hard=memory.available<50Mi,nodefs.available<1%,imagefs.available<1%"' \
      > /etc/rancher/k3s/config.yaml \
 && printf '%s\n' '[Unit]' 'Description=Kubernetes (k3s)' 'After=network-online.target' '' \
      '[Service]' 'Type=notify' 'ExecStartPre=-/sbin/modprobe br_netfilter' \
      'ExecStart=/usr/local/bin/k3s server' 'KillMode=process' 'Delegate=yes' \
      'LimitNOFILE=1048576' 'LimitNPROC=infinity' 'LimitCORE=infinity' 'TasksMax=infinity' \
      'TimeoutStartSec=0' 'Restart=always' 'RestartSec=5s' '' \
      '[Install]' 'WantedBy=multi-user.target' > /etc/systemd/system/k3s.service \
 && echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' > /etc/profile.d/kubeconfig.sh \
 && printf '%s\n' '# kubectl: autocompletar e o atalho k (só no bash)' \
      '[ -n "${BASH_VERSION:-}" ] || return 0' \
      'source <(kubectl completion bash 2>/dev/null)' \
      'alias k=kubectl' 'complete -o default -F __start_kubectl k' \
      > /etc/profile.d/kubectl-completion.sh

# Registry interna da loja (registry.loja.local): só responde a listagem de tags.
# As imagens da loja já estão no nó; uma tag que não existe dá "not found".
RUN printf '%s\n' 'mirrors:' '  registry.loja.local:' '    endpoint:' '      - "http://127.0.0.1:5000"' \
      > /etc/rancher/k3s/registries.yaml \
 && install -d /srv/registry \
 && printf '%s\n' '[Unit]' 'Description=Registry interna da Loja Exemplo' '' '[Service]' \
      'ExecStart=/usr/bin/python3 /usr/local/lib/registry-loja.py' 'Restart=always' '' \
      '[Install]' 'WantedBy=multi-user.target' > /etc/systemd/system/registry-loja.service

COPY registry-loja.py /usr/local/lib/
COPY k3s-lab-up /usr/local/bin/
