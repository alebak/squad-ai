#!/usr/bin/env bash
set -euo pipefail

# Squad AI — Install Script
# Manage your AI coding agent squad inside dev containers.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash

GITHUB_OWNER="alebak"
GITHUB_REPO="squad-ai"
BINARY_NAME="squad"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

success() { echo -e "${GREEN}[ok]${NC}      $*"; }
error()   { echo -e "${RED}[error]${NC}   $*" >&2; }

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux)  os="linux" ;;
        Darwin) os="darwin" ;;
        *)      error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)             error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}/${arch}"
}

get_latest_version() {
    local url="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest"
    local response tag

    response="$(curl -sfL "$url")" || { error "Failed to fetch latest release from GitHub"; exit 1; }
    tag="$(echo "$response" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"

    if [ -z "$tag" ]; then
        error "Could not determine latest version"
        exit 1
    fi

    echo "${tag#v}"
}

# Temp directory is global so the EXIT trap can clean it up.
TMPDIR=""
trap 'rm -rf "$TMPDIR"' EXIT

main() {
    echo -e "\n${CYAN}${BOLD}Squad AI — Installer${NC}\n"

    local platform version archive install_dir
    platform="$(detect_platform)"
    version="$(get_latest_version)"
    archive="squad-ai_${version}_${platform%/*}_${platform#*/}.tar.gz"
    install_dir="${HOME}/.local/bin"
    TMPDIR="$(mktemp -d)"

    success "Platform: ${platform}"
    success "Version: ${version}"

    echo -e "\nDownloading ${archive}..."
    if ! curl -sfL -o "${TMPDIR}/${archive}" \
        "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/v${version}/${archive}"; then
        error "Failed to download ${archive}"
        exit 1
    fi

    echo "Extracting..."
    tar -xzf "${TMPDIR}/${archive}" -C "$TMPDIR"

    mkdir -p "$install_dir"
    cp "${TMPDIR}/squad" "${install_dir}/squad"
    chmod +x "${install_dir}/squad"

    if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
        echo -e "\n${CYAN}Add to your shell profile:${NC}"
        echo "  export PATH=\"\$PATH:${install_dir}\""
    fi

    echo ""
    success "Installed squad to ${install_dir}/squad"
    echo -e "Run ${BOLD}squad${NC} to configure your AI coding agents.\n"
}

main "$@"
