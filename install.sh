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

# Détecter l'OS
OS="unknown"
case "$(uname -s)" in
    Linux*)     OS="linux";;
    Darwin*)    OS="macos";;
    CYGWIN*|MINGW*|MSYS*) OS="windows";;
esac

echo -e "${GREEN}Système détecté: $OS${NC}"

# Vérifier Docker
echo -e "\n${YELLOW}Vérification de Docker...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker n'est pas installé.${NC}"
    echo "Veuillez installer Docker: https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✓ Docker est installé${NC}"

# Construire l'image
echo -e "\n${YELLOW}Construction de l'image Docker...${NC}"
docker build -t torrent-aio:latest .
echo -e "${GREEN}✓ Image construite avec succès${NC}"

# Installation du script
echo -e "\n${YELLOW}Installation du script...${NC}"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

if [ "$OS" = "windows" ]; then
    echo -e "${YELLOW}Sur Windows, ajoutez le dossier 'scripts' à votre PATH${NC}"
    echo "Ou copiez scripts/torrent-aio.ps1 dans un dossier de votre PATH"
else
    cp scripts/torrent-aio.sh "$INSTALL_DIR/torrent-aio"
    chmod +x "$INSTALL_DIR/torrent-aio"
    echo -e "${GREEN}✓ Script installé dans $INSTALL_DIR/torrent-aio${NC}"
fi

# Résumé
echo -e "\n${GREEN}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              Installation terminée !                      ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Vérifier si le dossier d'installation est dans le PATH
if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
    echo -e "${YELLOW}Attention : le dossier $INSTALL_DIR n'est pas dans votre PATH.${NC}"
    echo -e "Ajoutez la ligne suivante à votre fichier de configuration de shell (ex: ~/.bashrc, ~/.zshrc) :"
    echo -e "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo -e "Puis rechargez votre shell ou ouvrez un nouveau terminal."
fi

echo "Utilisation:"
echo -e "  ${BLUE}torrent-aio process /chemin/vers/film.mkv${NC}"
echo ""
echo "Options:"
echo "  --help           Afficher l'aide"
echo "  --group NAME     Nom du groupe de release"
echo ""
echo "Documentation: https://github.com/metwurcht/torrent-all-in-one"
