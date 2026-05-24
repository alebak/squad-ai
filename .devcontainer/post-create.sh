#!/bin/bash
set -e

# Persist PATH for future shells
mkdir -p "$HOME/.local/bin"
if ! grep -qF '.local/bin' "$HOME/.bashrc" 2>/dev/null; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
fi
export PATH="$HOME/.local/bin:$PATH"

# # Install Node.js
# if ! command -v node &>/dev/null; then
#   curl -fsSL https://deb.nodesource.com/setup_22.x | sudo bash -
#   sudo apt-get install -y nodejs
# fi

# # Set up global npm variables in the home directory (persists on the volume)
# mkdir -p "$HOME/.npm-global"
# npm config set prefix "$HOME/.npm-global"
# if ! grep -qF '.npm-global/bin' "$HOME/.bashrc" 2>/dev/null; then
#   echo 'export PATH="$HOME/.npm-global/bin:$PATH"' >> "$HOME/.bashrc"
# fi
# export PATH="$HOME/.npm-global/bin:$PATH"

# Install Squad AI — manages all coding agents
if command -v squad &>/dev/null; then
  echo "✅ Squad AI already installed"
else
  echo "📦 Installing Squad AI..."
  curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
fi

# Launch Squad AI — first run shows TUI, subsequent runs install silently
echo "📦 Syncing agents via Squad AI..."
squad
