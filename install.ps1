# Script d'installation de Torrent All-In-One pour Windows
# Usage: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

# Configuration
$GITHUB_REPO = "metwurcht/torrent-all-in-one"
$VERSION = if ($env:TORRENT_AIO_VERSION) { $env:TORRENT_AIO_VERSION } else { "latest" }

# Couleurs
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

Write-Host ""
Write-ColorOutput Blue "╔═══════════════════════════════════════════════════════════╗"
Write-ColorOutput Blue "║           Torrent All-In-One - Installation               ║"
Write-ColorOutput Blue "╚═══════════════════════════════════════════════════════════╝"
Write-Host ""

# Détecter l'architecture
$ARCH = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-ColorOutput Green "Système détecté: windows/$ARCH"

# Vérifier que mediainfo est installé
Write-Host ""
Write-ColorOutput Yellow "Vérification de mediainfo..."
$mediaInfoPath = Get-Command mediainfo.exe -ErrorAction SilentlyContinue
if (-not $mediaInfoPath) {
    Write-ColorOutput Red "mediainfo n'est pas installé."
    Write-Host "Installez-le avec :"
    Write-ColorOutput Blue "  winget install MediaArea.MediaInfo"
    Write-Host "Ou téléchargez-le depuis: https://mediaarea.net/en/MediaInfo/Download/Windows"
    exit 1
}
Write-ColorOutput Green "✓ mediainfo est installé"

# Déterminer le répertoire d'installation
$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\torrent-aio"
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
}

Write-ColorOutput Yellow "Répertoire d'installation: $INSTALL_DIR"

# Télécharger le binaire
Write-Host ""
Write-ColorOutput Yellow "Téléchargement de torrent-aio.exe..."

$BINARY_NAME = "torrent-aio-windows-$ARCH.exe"
if ($VERSION -eq "latest") {
    $DOWNLOAD_URL = "https://github.com/$GITHUB_REPO/releases/latest/download/$BINARY_NAME"
} else {
    $DOWNLOAD_URL = "https://github.com/$GITHUB_REPO/releases/download/$VERSION/$BINARY_NAME"
}

$DEST_FILE = "$INSTALL_DIR\torrent-aio.exe"

try {
    # Utiliser WebClient pour le téléchargement avec barre de progression
    $webClient = New-Object System.Net.WebClient
    $webClient.DownloadFile($DOWNLOAD_URL, $DEST_FILE)
} catch {
    Write-ColorOutput Red "Erreur lors du téléchargement: $_"
    exit 1
}

Write-ColorOutput Green "✓ Téléchargement réussi !"

# Ajouter au PATH si nécessaire
Write-Host ""
Write-ColorOutput Yellow "Configuration du PATH..."

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$INSTALL_DIR*") {
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$currentPath;$INSTALL_DIR",
        "User"
    )
    Write-ColorOutput Green "✓ $INSTALL_DIR ajouté au PATH"
    Write-ColorOutput Yellow "⚠ Redémarrez votre terminal pour que les changements prennent effet"
} else {
    Write-ColorOutput Green "✓ Le répertoire est déjà dans le PATH"
}

Write-Host ""
Write-ColorOutput Green "✓ Installation réussie !"
Write-ColorOutput Green "Le binaire est installé dans: $INSTALL_DIR\torrent-aio.exe"

Write-Host ""
Write-ColorOutput Green "Vous pouvez maintenant utiliser:"
Write-ColorOutput Blue "  torrent-aio <fichier_video>"
Write-ColorOutput Blue "  torrent-aio --help"
