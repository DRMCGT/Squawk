#!/usr/bin/env sh
# Squawk installer — installs the latest (or pinned) Squawk release binary.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/DRMCGT/Squawk/main/install.sh | sh
#   SQUAWK_VERSION=v0.1.0 sh <(curl -fsSL .../install.sh)
#
# Environment overrides:
#   SQUAWK_VERSION   release tag to install (default: latest)
#   SQUAWK_INSTALL_DIR   install directory (default: ~/.local/bin)
#   SQUAWK_REPO      GitHub repo (default: DRMCGT/Squawk)
set -eu

REPO="${SQUAWK_REPO:-DRMCGT/Squawk}"
VERSION="${1:-${SQUAWK_VERSION:-latest}}"
INSTALL_DIR="${SQUAWK_INSTALL_DIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux|darwin) ;;
  *)
    echo "error: unsupported operating system: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required to install Squawk" >&2
  exit 1
fi

asset="squawk-${os}-${arch}.tar.gz"

if [ "$VERSION" = "latest" ]; then
  base="https://github.com/${REPO}/releases/latest/download"
else
  base="https://github.com/${REPO}/releases/download/${VERSION}"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "squawk: downloading ${base}/${asset}"
curl -fsSL -o "$tmp/${asset}" "${base}/${asset}"

echo "squawk: verifying checksum"
curl -fsSL -o "$tmp/SHA256SUMS.txt" "${base}/SHA256SUMS.txt"
(
  cd "$tmp"
  if command -v sha256sum >/dev/null 2>&1; then
    grep "$asset" SHA256SUMS.txt | sha256sum -c -
  elif command -v shasum >/dev/null 2>&1; then
    grep "$asset" SHA256SUMS.txt | shasum -a 256 -c -
  else
    echo "warning: no sha256 verification tool found; skipping checksum" >&2
  fi
)

echo "squawk: installing to ${INSTALL_DIR}"
mkdir -p "$INSTALL_DIR"
tar -xzf "$tmp/${asset}" -C "$INSTALL_DIR" squawk
chmod +x "$INSTALL_DIR/squawk"

if ! printf ':%s:' "$PATH" | grep -q ":${INSTALL_DIR}:"; then
  echo "note: ${INSTALL_DIR} is not on your PATH."
  shell="$(basename "${SHELL:-unknown}")"
  case "$shell" in
    zsh) rc="$HOME/.zshrc" ;;
    *) rc="$HOME/.bashrc" ;;
  esac
  echo "      Add it with:"
  echo "      echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ${rc} && source ${rc}"
fi

"$INSTALL_DIR/squawk" version