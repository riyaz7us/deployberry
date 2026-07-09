#!/bin/bash

# DeployBerry Quick Installer Script

set -e

# Ensure the script is run as root
if [ "$EUID" -ne 0 ]; then
  echo "Error: This installer must be run as root (sudo)."
  exit 1
fi

echo "======================================"
echo "DeployBerry Quick Installer"
echo "======================================"
echo

# Determine architecture
ARCH=$(uname -m)
BINARY_URL="https://github.com/riyaz7us/deployberry/releases/latest/download/deployberry-linux-amd64"

if [ "$ARCH" != "x86_64" ]; then
  echo "Warning: Only x86_64 (amd64) architecture is officially pre-compiled."
  echo "Attempting to download default linux-amd64 binary..."
fi

TEMP_BIN="/tmp/deployberry_install_bin"

echo "Downloading the latest DeployBerry binary..."
curl -fsSL "$BINARY_URL" -o "$TEMP_BIN"
chmod +x "$TEMP_BIN"

echo "Running DeployBerry system installer..."
"$TEMP_BIN" install

# Clean up installer temp binary
rm -f "$TEMP_BIN"

echo "--------------------------------------"
echo "Creating Administrator Account"
echo "--------------------------------------"
read -p "Enter admin username: " admin_user
read -s -p "Enter admin password: " admin_pass
echo

if [ -z "$admin_user" ] || [ -z "$admin_pass" ]; then
  echo "Error: Username and password cannot be empty. Admin registration skipped."
  echo "You can register later by running: sudo deployberry register <username> <password>"
else
  deployberry register "$admin_user" "$admin_pass"
fi

echo
echo "======================================"
echo "Setup Complete!"
echo "DeployBerry is now running on http://your-server-ip:7717"
echo "======================================"
