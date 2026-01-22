# Torrent All-In-One - Installation (Windows PowerShell)

Write-Host "" -ForegroundColor Blue
Write-Host "╔═══════════════════════════════════════════════════════════╗"
Write-Host "║           Torrent All-In-One - Installation               ║"
Write-Host "╚═══════════════════════════════════════════════════════════╝"
Write-Host ""

# Détecter l'OS
$OS = "windows"
Write-Host "Système détecté: $OS" -ForegroundColor Green

# Vérifier Docker
Write-Host "\nVérification de Docker..." -ForegroundColor Yellow
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "Docker n'est pas installé." -ForegroundColor Red
    Write-Host "Veuillez installer Docker: https://docs.docker.com/get-docker/"
    exit 1
}
Write-Host "✓ Docker est installé" -ForegroundColor Green

# Construire l'image
Write-Host "\nConstruction de l'image Docker..." -ForegroundColor Yellow
try {
    docker build -t torrent-aio:latest .
    Write-Host "✓ Image construite avec succès" -ForegroundColor Green
} catch {
    Write-Host "Erreur lors de la construction de l'image Docker." -ForegroundColor Red
    exit 1
}

# Installation du script
Write-Host "\nInstallation du script..." -ForegroundColor Yellow
$scriptSource = "scripts/torrent-aio.ps1"
$installDir = "$env:USERPROFILE\bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}
Copy-Item $scriptSource "$installDir\torrent-aio.ps1" -Force
Write-Host "✓ Script installé dans $installDir\torrent-aio.ps1" -ForegroundColor Green

# Résumé
Write-Host "\n"
Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║              Installation terminée !                      ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

# Vérifier si le dossier d'installation est dans le PATH
if (-not ($env:PATH -split ';' | Where-Object { $_ -eq $installDir })) {
    Write-Host "Attention : le dossier $installDir n'est pas dans votre PATH." -ForegroundColor Yellow
    Write-Host "Ajoutez la ligne suivante à votre profil PowerShell (ex: $PROFILE) :"
    Write-Host "  $env:PATH += ';$installDir'"
    Write-Host "Puis rechargez votre session ou ouvrez un nouveau terminal."
}

Write-Host "Utilisation:"
Write-Host "  torrent-aio.ps1 process C:\chemin\vers\film.mkv" -ForegroundColor Blue
Write-Host ""
Write-Host "Options:"
Write-Host "  --help           Afficher l'aide"
Write-Host "  --group NAME     Nom du groupe de release"
Write-Host "  --no-rename      Ne pas renommer le fichier"
Write-Host "  --skip-torrent   Ne pas générer le fichier torrent"
Write-Host ""
Write-Host "Configuration:"
Write-Host "  Créez %USERPROFILE%\.config\torrent-aio.yml pour définir des valeurs par défaut"
Write-Host "  Exemple:"
Write-Host "    group_name: \"MW\""
Write-Host "    skip_torrent: false"
Write-Host ""
Write-Host "Documentation: https://github.com/metwurcht/torrent-all-in-one"
