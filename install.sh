#!/bin/bash
#
# Script d'installation de Torrent All-In-One
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           Torrent All-In-One - Installation               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Configuration
GITHUB_REPO="metwurcht/torrent-all-in-one"
VERSION="${TORRENT_AIO_VERSION:-latest}"

# Détecter l'OS et l'architecture
OS="unknown"
ARCH="unknown"
case "$(uname -s)" in
    Linux*)     OS="linux";;
    Darwin*)    OS="darwin";;
    *)          echo -e "${RED}OS non supporté${NC}"; exit 1;;
esac

case "$(uname -m)" in
    x86_64)     ARCH="amd64";;
    aarch64|arm64) ARCH="arm64";;
    *)          echo -e "${RED}Architecture non supportée${NC}"; exit 1;;
esac

echo -e "${GREEN}Système détecté: $OS/$ARCH${NC}"

# Déterminer le répertoire d'installation
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

echo -e "${YELLOW}Répertoire d'installation: $INSTALL_DIR${NC}"

# Télécharger le binaire
echo -e "\n${YELLOW}Téléchargement de torrent-aio...${NC}"

BINARY_NAME="torrent-aio-${OS}-${ARCH}"
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}"
else
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}"
fi

TMP_FILE="/tmp/torrent-aio"
if command -v curl &> /dev/null; then
    curl -L -o "$TMP_FILE" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -O "$TMP_FILE" "$DOWNLOAD_URL"
else
    echo -e "${RED}Erreur: curl ou wget est requis${NC}"
    exit 1
fi

# Installer le binaire
echo -e "${YELLOW}Installation...${NC}"
chmod +x "$TMP_FILE"
mv "$TMP_FILE" "$INSTALL_DIR/torrent-aio"

echo -e "${GREEN}✓ Installation réussie !${NC}"
echo -e "${GREEN}Le binaire est installé dans: $INSTALL_DIR/torrent-aio${NC}"

# Vérifier si le répertoire est dans le PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "\n${YELLOW}⚠ Attention: $INSTALL_DIR n'est pas dans votre PATH${NC}"
    echo -e "Ajoutez cette ligne à votre ~/.bashrc ou ~/.zshrc :"
    echo -e "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo -e "\n${GREEN}Vous pouvez maintenant utiliser:${NC}"
echo -e "  ${BLUE}torrent-aio <fichier_video>${NC}"
echo -e "  ${BLUE}torrent-aio --help${NC}"
