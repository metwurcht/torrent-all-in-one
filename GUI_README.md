# Torrent All-In-One GUI

Interface graphique pour Torrent All-In-One construite avec Wails et Svelte.

## Prérequis

- Go 1.21+
- Node.js 18+
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Dépendances système (Linux)

```bash
# Ubuntu/Debian
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch
sudo pacman -S gtk3 webkit2gtk
```

## Installation

1. Installer les dépendances frontend:

```bash
cd frontend
npm install
```

2. Générer les bindings Wails:

```bash
wails generate module
```

## Développement

Lancer en mode développement (avec hot-reload):

```bash
wails dev
```

## Build

### Build pour votre plateforme:

```bash
wails build
```

### Build multi-plateforme:

```bash
# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64

# macOS
wails build -platform darwin/universal
```

Les binaires seront dans le dossier `build/bin/`.

## Architecture

```
cmd/torrent-aio-gui/
  └── main.go              # Point d'entrée Wails

internal/gui/
  ├── app.go              # Service principal (exposé au frontend)
  ├── reporter.go         # Reporter pour la progression
  └── prompter.go         # Prompter pour les interactions

frontend/
  ├── src/
  │   ├── App.svelte     # Interface principale
  │   ├── main.js        # Point d'entrée JS
  │   └── style.css      # Styles
  ├── index.html
  ├── package.json
  └── vite.config.js
```

## Fonctionnalités

- ✅ Sélection de fichier via dialogue natif
- ✅ Configuration des options (groupe, output, etc.)
- ✅ Progression en temps réel
- ✅ Affichage des résultats
- 🚧 Sélection du type de source (à implémenter)
- 🚧 Sélection du film TMDB (à implémenter)

## Utilisation

1. Cliquer sur "Parcourir" pour sélectionner un fichier vidéo
2. (Optionnel) Configurer le nom du groupe et le dossier de sortie
3. Cocher les options désirées
4. Cliquer sur "Traiter"
5. Suivre la progression dans les logs
6. Les fichiers générés sont affichés dans la section "Résultat"
