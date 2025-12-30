#!/bin/bash
#
# Torrent All-In-One - Script wrapper pour Docker
# Ce script permet d'utiliser torrent-aio facilement via Docker
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="${TORRENT_AIO_IMAGE:-torrent-aio:latest}"
CONTAINER_NAME="torrent-aio-run"

# Vérifier que Docker est disponible
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Erreur: Docker n'est pas installé${NC}"
    exit 1
fi

# Construire l'image si elle n'existe pas
if ! docker image inspect "$IMAGE_NAME" &> /dev/null; then
    echo -e "${YELLOW}Image Docker non trouvée, construction en cours...${NC}"
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    docker build -t "$IMAGE_NAME" "$SCRIPT_DIR"
fi

# Déterminer le chemin du fichier
FILE_PATH=""
ARGS=()

for arg in "$@"; do
    if [ -f "$arg" ]; then
        FILE_PATH="$arg"
        ABS_PATH="$(cd "$(dirname "$arg")" && pwd)/$(basename "$arg")"
        MOUNT_DIR="$(dirname "$ABS_PATH")"
        FILE_NAME="$(basename "$arg")"
        ARGS+=("/data/$FILE_NAME")
    else
        ARGS+=("$arg")
    fi
done

# Exécuter le conteneur
if [ -n "$FILE_PATH" ]; then
    docker run --rm -it \
        -e TORRENT_AIO_TRACKER_URL="${TRACKER_URL:-}" \
        -e TORRENT_AIO_GROUP_NAME="${GROUP_NAME:-TORRENT-AIO}" \
        -v "$MOUNT_DIR:/data" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" "${ARGS[@]}"
else
    docker run --rm -it \
        -e TORRENT_AIO_TRACKER_URL="${TRACKER_URL:-}" \
        -e TORRENT_AIO_GROUP_NAME="${GROUP_NAME:-TORRENT-AIO}" \
        -v "$(pwd):/data" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" "${ARGS[@]}"
fi
