<#
.SYNOPSIS
    Torrent All-In-One - Script wrapper pour Docker (PowerShell)
.DESCRIPTION
    Ce script permet d'utiliser torrent-aio facilement via Docker sur Windows
.PARAMETER Args
    Arguments à passer à torrent-aio
#>

param(
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

$ErrorActionPreference = "Stop"

# Configuration
$ImageName = if ($env:TORRENT_AIO_IMAGE) { $env:TORRENT_AIO_IMAGE } else { "torrent-aio:latest" }
$ContainerName = "torrent-aio-run"

# Vérifier que Docker est disponible
try {
    docker --version | Out-Null
} catch {
    Write-Host "Erreur: Docker n'est pas installé" -ForegroundColor Red
    exit 1
}

# Vérifier si l'image existe
$imageExists = docker image inspect $ImageName 2>$null
if (-not $imageExists) {
    Write-Host "Image Docker non trouvée, construction en cours..." -ForegroundColor Yellow
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $projectDir = Split-Path -Parent $scriptDir
    docker build -t $ImageName $projectDir
}

# Préparer les arguments
$dockerArgs = @()
$mountDir = $PWD.Path
$processedArgs = @()

foreach ($arg in $Arguments) {
    if (Test-Path $arg -PathType Leaf) {
        $fullPath = (Resolve-Path $arg).Path
        $mountDir = Split-Path -Parent $fullPath
        $fileName = Split-Path -Leaf $fullPath
        $processedArgs += "/data/$fileName"
    } else {
        $processedArgs += $arg
    }
}

# Convertir le chemin Windows en format Docker (pour Docker Desktop)
$mountDirDocker = $mountDir -replace '\\', '/' -replace '^([A-Za-z]):', '/$1'

# Exécuter le conteneur
$dockerCommand = @(
    "run", "--rm", "-it",
    "-e", "TORRENT_AIO_GROUP_NAME=$($env:GROUP_NAME ?? 'TORRENT-AIO')",
    "-v", "${mountDir}:/data",
    "--name", $ContainerName,
    $ImageName
) + $processedArgs

& docker @dockerCommand
