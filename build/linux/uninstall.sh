#!/usr/bin/env bash
# Removes everything install.sh created. Your data (the SQLite database
# under ~/.config/faster) is left untouched.
set -euo pipefail

rm -f "$HOME/.local/bin/faster"
rm -f "$HOME/.local/share/icons/hicolor/256x256/apps/faster.png"
rm -f "$HOME/.local/share/applications/faster.desktop"
rm -f "$HOME/Desktop/faster.desktop"

command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$HOME/.local/share/applications" || true

echo "Faster has been uninstalled. Your data in ~/.config/faster was left in place —"
echo "delete that folder too if you want to remove it."
