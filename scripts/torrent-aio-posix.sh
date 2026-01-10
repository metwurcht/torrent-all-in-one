#!/bin/sh
#
# Torrent All-In-One - Script wrapper pour Docker (POSIX sh compatible)
# Compatible avec sh, dash, busybox ash, etc.
#

set -e

# Configuration
IMAGE_NAME="${TORRENT_AIO_IMAGE:-torrent-aio:latest}"
CONTAINER_NAME="torrent-aio-run"

# Vérifier que Docker est disponible
if ! command -v docker >/dev/null 2>&1; then
    echo "Erreur: Docker n'est pas installé" >&2
    exit 1
fi

# Construire l'image si elle n'existe pas
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
    echo "Image Docker non trouvée, construction en cours..."
    SCRIPT_DIR="$(dirname "$0")"
    PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
    docker build -t "$IMAGE_NAME" "$PROJECT_DIR"
fi

# Déterminer le chemin du fichier
FILE_PATH=""
NEW_ARGS=""

for arg in "$@"; do
    if [ -f "$arg" ]; then
        FILE_PATH="$arg"
        # Obtenir le chemin absolu
        ABS_DIR="$(cd "$(dirname "$arg")" && pwd)"
        FILE_NAME="$(basename "$arg")"
        MOUNT_DIR="$ABS_DIR"
        NEW_ARGS="$NEW_ARGS /data/$FILE_NAME"
    else
        NEW_ARGS="$NEW_ARGS $arg"
    fi
done

# Exécuter le conteneur
if [ -n "$FILE_PATH" ]; then
    docker run --rm -it \
        -e TORRENT_AIO_GROUP_NAME="${GROUP_NAME:-TORRENT-AIO}" \
        -v "$MOUNT_DIR:/data" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" $NEW_ARGS
else
    docker run --rm -it \
        -e TORRENT_AIO_GROUP_NAME="${GROUP_NAME:-TORRENT-AIO}" \
        -v "$(pwd):/data" \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" $NEW_ARGS
fi
