# Torrent All-In-One

🎬 Outil CLI et GUI pour préparer des releases de films : identification TMDB (scraping), analyse technique, génération NFO et création de torrent.

> ⚠️ **Prérequis important** : [MediaInfo](https://mediaarea.net/en/MediaInfo) doit être installé sur votre système pour que l'outil fonctionne correctement.

## ✨ Fonctionnalités

- **Identification automatique** : Recherche le film sur TMDB via scraping (aucune clé API requise)
- **Sélection interactive** : Choix parmi les résultats ou recherche manuelle / ID direct
- **Analyse technique** : Extraction des métadonnées via MediaInfo
- **Renommage automatique** : Convention de nommage warez (Titre.Année.Résolution.Source.Codec-GROUPE)
- **Génération NFO** : Fichier NFO avec infos film et techniques
- **Présentation BBCode** : Résumé formaté pour forums
- **Création torrent** : Génération du fichier .torrent
- **Interface graphique** : GUI avec Wails (optionnel)

## 🚀 Installation

### ⚠️ Prérequis : MediaInfo

Avant d'installer Torrent All-In-One, assurez-vous d'avoir MediaInfo installé :

**Linux (Debian/Ubuntu)** :

```bash
sudo apt install mediainfo
```

**Linux (Fedora)** :

```bash
sudo dnf install mediainfo
```

**macOS** :

```bash
brew install media-info
```

**Windows** :

```powershell
winget install MediaArea.MediaInfo
```

Ou téléchargez depuis [mediaarea.net](https://mediaarea.net/en/MediaInfo/Download)

---

## 🖥️ Interface en Ligne de Commande (CLI)

### Installation du CLI

#### Option 1 : Installation automatique (recommandé)

**Linux/macOS** :

```bash
curl -sSL https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.sh | bash
```

Ou avec wget :

```bash
wget -qO- https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.sh | bash
```

**Windows** :

```powershell
irm https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.ps1 | iex
```

Ou téléchargez et exécutez [install.ps1](install.ps1) :

```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

#### Option 2 : Installation manuelle depuis les releases

1. Téléchargez le binaire correspondant à votre système depuis les [releases GitHub](https://github.com/metwurcht/torrent-all-in-one/releases) :

**Linux** :

- `torrent-aio-linux-amd64` - Linux x86_64
- `torrent-aio-linux-arm64` - Linux ARM64

**macOS** :

- `torrent-aio-darwin-amd64` - macOS Intel
- `torrent-aio-darwin-arm64` - macOS Apple Silicon

**Windows** :

- `torrent-aio-windows-amd64.exe` - Windows 64-bit
- `torrent-aio-windows-arm64.exe` - Windows ARM64

2. Installez le binaire :

**Linux/macOS** :

```bash
# Télécharger (remplacez VERSION et PLATFORM par les valeurs appropriées)
curl -L -o torrent-aio https://github.com/metwurcht/torrent-all-in-one/releases/latest/download/torrent-aio-linux-amd64

# Rendre exécutable
chmod +x torrent-aio

# Installer dans le PATH
sudo mv torrent-aio /usr/local/bin/
```

**Windows** :

- Téléchargez le `.exe` depuis les releases
- Placez-le dans un dossier de votre choix
- Ajoutez ce dossier à votre PATH

### Utilisation du CLI

#### Commande de base

```bash
torrent-aio /chemin/vers/film.mkv
```

#### Options avancées

```bash
torrent-aio film.mkv \
  --tracker "http://tracker.example.com/announce" \
  --group "MONGROUPE" \
  --output /chemin/sortie \
  --no-rename          # Ne pas renommer le fichier
  --skip-torrent       # Ne pas générer le fichier torrent
```

#### Fichier de configuration

Créez `~/.config/torrent-aio.yml` pour définir des valeurs par défaut :

```yaml
group_name: "MONGROUPE"
tracker_url: "http://tracker.example.com/announce"
```

---

## 🎨 Interface Graphique (GUI)

### Présentation

L'interface graphique `torrent-aio-gui` offre une expérience utilisateur visuelle et intuitive pour préparer vos releases. Elle inclut toutes les fonctionnalités du CLI dans une interface moderne et facile à utiliser.

### Installation du GUI

1. Téléchargez la version GUI correspondant à votre système depuis les [releases GitHub](https://github.com/metwurcht/torrent-all-in-one/releases) :

**Linux** :

- `torrent-aio-gui-linux-amd64` - Linux x86_64

**macOS** :

- `torrent-aio-gui-darwin-universal.zip` - macOS Intel et Apple Silicon
  - Décompressez le fichier et déplacez l'application `.app` dans votre dossier Applications

**Windows** :

- `torrent-aio-gui-windows-amd64.exe` - Windows 64-bit

2. Lancez l'application :

**Linux** :

```bash
chmod +x torrent-aio-gui-linux-amd64
./torrent-aio-gui-linux-amd64
```

**macOS** :

- Double-cliquez sur l'application dans le dossier Applications
- Si macOS bloque l'application : Préférences Système → Sécurité → Autoriser

**Windows** :

- Double-cliquez sur le fichier `.exe`
- **⚠️ Avertissement antivirus** : Windows Defender peut bloquer l'exécutable (faux positif). Voir la section [Problèmes courants](#-problèmes-courants) ci-dessous.

### ⚠️ Windows : Faux positifs antivirus

Les antivirus (notamment Windows Defender) détectent parfois le fichier `.exe` comme suspect. **C'est un faux positif** car :

- L'application n'est pas signée numériquement (signature coûteuse ~500$/an)
- Les binaires Go/Wails sont souvent détectés à tort
- L'application est récente et peu téléchargée

**Solutions** :

1. **Autoriser l'application dans Windows Defender** :
   - Ouvrez "Sécurité Windows" → "Protection contre les virus et menaces"
   - Cliquez sur "Gérer les paramètres"
   - Descendez à "Exclusions" → "Ajouter ou supprimer des exclusions"
   - Ajoutez le fichier `.exe` ou le dossier contenant l'application

2. **Télécharger depuis GitHub** :
   - Assurez-vous de télécharger uniquement depuis les [releases officielles](https://github.com/metwurcht/torrent-all-in-one/releases)
   - Vérifiez que l'URL commence par `github.com/metwurcht/torrent-all-in-one`

**Le code source est open-source** : Vous pouvez vérifier qu'il n'y a rien de malveillant dans le dépôt GitHub.

### Utilisation du GUI

1. Lancez l'application
2. Sélectionnez votre fichier vidéo
3. Suivez les étapes guidées dans l'interface
4. L'application génère automatiquement tous les fichiers nécessaires

---

## 🔧 Workflow

1. **Analyse parallèle** : Le fichier est analysé en arrière-plan pendant la recherche TMDB
2. **Recherche TMDB** : Les mots-clés sont extraits du nom de fichier (scraping web)
3. **Sélection** : Choisissez le bon film dans la liste ou :
   - Tapez `0` pour une nouvelle recherche
   - Entrez `id:12345` pour utiliser un ID TMDB directement
4. **Génération** :
   - Le fichier est renommé selon la convention warez
   - Un fichier NFO est créé
   - Le résumé bbcode est affiché dans la console
   - Le fichier torrent est généré

## 🏗️ Architecture

```
torrent-all-in-one/
├── cmd/
│   ├── torrent-aio/      # CLI - Point d'entrée en ligne de commande
│   └── torrent-aio-gui/  # GUI - Interface graphique (Wails)
├── internal/
│   ├── cli/              # Commandes Cobra
│   ├── gui/              # Logique GUI
│   ├── tmdb/             # Client TMDB (scraping web)
│   ├── mediainfo/        # Analyse fichiers vidéo
│   ├── nfo/              # Génération NFO
│   ├── renamer/          # Renommage warez
│   ├── presenter/        # Génération présentation BBCode
│   ├── torrent/          # Génération torrent
│   └── ui/               # Interface utilisateur
└── frontend/             # Interface Svelte (GUI)
```

## 📝 Licence

MIT

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une PR.
