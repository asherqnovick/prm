#!/bin/bash
# install.sh

set -e

BINARY_NAME="prm"
INSTALL_DIR="/usr/local/bin"
BASE_URL="https://github.com/asherqnovick/prm/releases/download/v0.0.1"

OS=$(uname -s)
case "$OS" in
    Darwin)
        BINARY_FILE="prm_darwin_universal"
        ;;
    Linux)
        BINARY_FILE="prm_linux_amd64"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

DOWNLOAD_URL="${BASE_URL}/${BINARY_FILE}"

curl -sSL -o "/tmp/${BINARY_NAME}" "${DOWNLOAD_URL}"
chmod +x "/tmp/${BINARY_NAME}"
sudo mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

echo "${BINARY_NAME} successfully installed"
