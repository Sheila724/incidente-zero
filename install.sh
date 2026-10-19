#!/usr/bin/env sh
# Instalador do Incidente Zero
#
#   curl -fsSL https://raw.githubusercontent.com/Sheila724/incidente-zero/main/install.sh | sh
#
# Variáveis opcionais:
#   IZ_VERSION=v0.1.0          versão específica (padrão: a mais recente)
#   IZ_INSTALL_DIR=/algum/dir  onde instalar o binário
set -eu

REPO="Sheila724/incidente-zero"
VERSION="${IZ_VERSION:-latest}"

info() { printf '\033[36m▶\033[0m %s\n' "$*"; }
warn() { printf '\033[35m!\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

# ---------- plataforma ----------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  *) die "Sistema não suportado: $os. No Windows, rode este instalador dentro do WSL2." ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "Arquitetura não suportada: $arch" ;;
esac

# ---------- ferramentas ----------
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1" -o "$2"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -q "$1" -O "$2"; }
else
  die "Preciso do curl ou do wget para baixar o iz."
fi

if command -v sha256sum >/dev/null 2>&1; then
  sha256() { sha256sum "$1" | cut -d' ' -f1; }
elif command -v shasum >/dev/null 2>&1; then
  sha256() { shasum -a 256 "$1" | cut -d' ' -f1; }
else
  die "Preciso do sha256sum ou do shasum para conferir o download."
fi

# ---------- download ----------
asset="iz_${os}_${arch}.tar.gz"
if [ -n "${IZ_BASE_URL:-}" ]; then
  base="$IZ_BASE_URL"   # espelho próprio ou testes
elif [ "$VERSION" = "latest" ]; then
  base="https://github.com/${REPO}/releases/latest/download"
else
  base="https://github.com/${REPO}/releases/download/${VERSION}"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

info "Baixando ${asset} (${VERSION})"
fetch "${base}/${asset}" "${tmp}/${asset}" || die "Não consegui baixar ${base}/${asset}"
fetch "${base}/checksums.txt" "${tmp}/checksums.txt" || die "Não consegui baixar o checksums.txt"

expected=$(grep " ${asset}\$" "${tmp}/checksums.txt" | cut -d' ' -f1)
[ -n "$expected" ] || die "O ${asset} não está no checksums.txt"
actual=$(sha256 "${tmp}/${asset}")
[ "$expected" = "$actual" ] || die "Checksum não confere: o download pode estar corrompido."
info "Checksum conferido"

tar -xzf "${tmp}/${asset}" -C "$tmp" iz

# ---------- instalação ----------
dir="${IZ_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ]; then
    dir=/usr/local/bin
  else
    dir="${HOME}/.local/bin"
  fi
fi
mkdir -p "$dir"
install -m 0755 "${tmp}/iz" "${dir}/iz" 2>/dev/null || { cp "${tmp}/iz" "${dir}/iz" && chmod 0755 "${dir}/iz"; }
info "Instalado em ${dir}/iz"

case ":${PATH}:" in
  *":${dir}:"*) ;;
  *) warn "${dir} não está no seu PATH. Adicione com: export PATH=\"${dir}:\$PATH\"" ;;
esac

# ---------- próximos passos ----------
if ! command -v docker >/dev/null 2>&1; then
  warn "Docker não encontrado. Instale o Docker antes de usar o Incidente Zero: https://docs.docker.com/get-docker/"
fi

"${dir}/iz" version || true
cat <<'EOF'

Próximos passos:
  iz pull      # baixa a imagem base dos labs
  iz doctor    # confere se está tudo pronto
  iz ui        # abre a interface em http://127.0.0.1:8080

Em Linux, instale o Sysbox para isolar os labs sem --privileged:
  https://github.com/nestybox/sysbox
EOF
