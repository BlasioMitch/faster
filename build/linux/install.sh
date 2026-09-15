#!/usr/bin/env bash
# Installs Faster for the current user: no sudo needed, nothing touches
# system directories. Adds it to your application menu (the Linux
# equivalent of the Start Menu) and gives you a proper icon you can drag to
# your Desktop.
#
# Run this from the extracted release folder (it expects `faster`,
# `faster.desktop` and `faster.png` next to it):
#   tar -xzf faster_*_linux_amd64.tar.gz
#   ./install.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# A single self-contained binary (the frontend is embedded in it), so it
# installs straight into ~/.local/bin — deliberately NOT
# ~/.local/share/faster, which is where WebKitGTK keeps its own per-app
# browser-engine cache (HSTS store, media keys, storage) once the app has
# run; sharing that folder with the installed binary is confusing even
# though it's harmless.
INSTALL_PATH="$HOME/.local/bin/faster"
ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
DESKTOP_DIR="$HOME/.local/share/applications"
DESKTOP_FILE="$DESKTOP_DIR/faster.desktop"

if [ ! -f "$SCRIPT_DIR/faster" ]; then
  echo "Error: expected to find 'faster' next to this script (in $SCRIPT_DIR)." >&2
  exit 1
fi

mkdir -p "$HOME/.local/bin" "$ICON_DIR" "$DESKTOP_DIR"

cp "$SCRIPT_DIR/faster" "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

if [ -f "$SCRIPT_DIR/faster.png" ]; then
  cp "$SCRIPT_DIR/faster.png" "$ICON_DIR/faster.png"
fi

if [ -f "$SCRIPT_DIR/faster.desktop" ]; then
  sed "s|__INSTALL_PATH__|$INSTALL_PATH|" "$SCRIPT_DIR/faster.desktop" > "$DESKTOP_FILE"
else
  cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=Faster
Comment=Offline fasting tracker
Exec=$INSTALL_PATH
Icon=faster
Terminal=false
Categories=Utility;
EOF
fi
chmod +x "$DESKTOP_FILE"

command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$DESKTOP_DIR" || true
command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f "$HOME/.local/share/icons/hicolor" 2>/dev/null || true

echo "Installed. Faster should now show up in your application menu (it can take"
echo "a moment, or a logout/login, to appear on some desktops)."
echo
echo "For a desktop icon, copy the entry there and mark it trusted:"
echo "  cp \"$DESKTOP_FILE\" \"\$HOME/Desktop/\" && gio set \"\$HOME/Desktop/faster.desktop\" metadata::trusted true"
echo "(GNOME/Nautilus requires that 'trusted' flag before a desktop launcher will run;"
echo "other desktop environments may just need you to right-click it and choose"
echo "'Allow Launching' instead.)"
echo
echo "To remove Faster later, run: $SCRIPT_DIR/uninstall.sh"
